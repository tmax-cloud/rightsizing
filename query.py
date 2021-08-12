import asyncio
from datetime import datetime
import logging
import os

from fastapi_cache.decorator import cache
from fastapi_cache.coder import PickleCoder
import httpx
import pandas as pd
import numpy as np

import constants
from models import AnalysisQuery, QueryParams
from utils import redis_key_builder

logger = logging.getLogger(__name__)


async def query_prometheus(client: httpx.AsyncClient, url: str, query: dict) -> dict:
    """
    Url에 해당하는 prometheus에 쿼리 요청하고 쿼리 결과를 저장하고 있는 리스트 리턴
    :param client: httpx.AsyncClient client
    :param url: The url of prometheus (ex. http://127.0.0.1:9000)
    :param query: The query dictionary (ex. {query: "container_cpu_usage_total{pod="test"}[1d]"}
    :return: The result of query
    """
    if not url.endswith(constants.QUERY_ENDPOINT):
        if url.endswith('/'):
            url = url[:-1]
        url = url + constants.QUERY_ENDPOINT

    resp = await client.get(url, params=query, timeout=60)
    if resp.status_code != 200:
        return None
    data = resp.json()
    if 'data' not in data:
        return None
    return data['data']['result']


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
async def query_and_analyze_pod(query: AnalysisQuery) -> QueryParams:
    """Query and analyze kubernetes pod.

    :param namespace: The namespace of kubernetes object to identify
    :param name: The name of kubernetes object
    :return: pd.DataFrame
    """
    url = os.getenv("HOST")
    namespace = query.namespace
    name = query.name

    def query_fmt(field):
        if namespace:
            q = f'{field}{{pod="{name}", namespace="{namespace}", prometheus=""}}[{constants.QUERY_TERM}]'
        else:
            q = f'{field}{{pod="{name}",container!="{constants.PAUSE_CONTAINER}"}}[{constants.QUERY_TERM}'
        return {'query': q}

    async with httpx.AsyncClient() as client:
        tasks = [query_prometheus(client, url, query_fmt(field)) for field in constants.FIELDS]
        result = await asyncio.gather(*tasks)

    total_df = pd.DataFrame()
    for records in result:
        for metric in records:
            field = metric['metric'].get('__name__', None)
            name = metric['metric'].get('name', None)

            times, values = map(list, zip(*metric['values']))
            if 'ds' not in total_df:
                total_df['ds'] = pd.date_range(start=datetime.fromtimestamp(times[0]), periods=len(times), freq='30s')

            values = np.array(values, dtype=np.float64)
            total_df[field] = values

    total_df = total_df.fillna(method='ffill')
    total_df = total_df.set_index('ds')

    return QueryParams(query=query, data=total_df)
