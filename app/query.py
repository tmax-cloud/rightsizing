import asyncio
from datetime import datetime
import time

from fastapi_cache.decorator import cache
from fastapi_cache.coder import PickleCoder
import httpx
import numpy as np
import pandas as pd

import constants
from .models import AnalysisQuery, QueryParams
from .utils import redis_key_builder


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

    resp = await client.get(url, params=query)
    if resp.status_code != 200:
        return None
    data = resp.json()
    if 'data' not in data:
        return None
    return data['data']['result']


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
async def query_and_analyze_pod(query: AnalysisQuery) -> QueryParams:
    """Query and analyze kubernetes pod.

    :param url: The url of prometheus server (ex. https://console.tmaxcloud.com/api/grafana/api/datasources/proxy/1)
    :param namespace: The namespace of kubernetes object to identify
    :param name: The name of kubernetes object
    :return: pd.DataFrame
    """
    url = query.url
    namespace = query.namespace
    name = query.name

    def query_fmt(field):
        if namespace:
            q = f'{field}{{pod="{name}",container!="{constants.PAUSE_CONTAINER}", ' \
                    f'namespace="{namespace}"}}[{constants.QUERY_TERM}]'
        else:
            q = f'{field}{{pod="{name}",container!="{constants.PAUSE_CONTAINER}"}}[{constants.QUERY_TERM}'
        return {'query': q}

    async with httpx.AsyncClient() as client:
        tasks = [query_prometheus(client, url, query_fmt(field)) for field in constants.FIELDS]
        result = await asyncio.gather(*tasks)
    series = dict()
    timestamps = None
    # for result in task_chord:
    for metrics in result:
        for container_metric in metrics:
            field = container_metric['metric'].get('__name__', None)
            pod = container_metric['metric'].get('pod', None)
            name = container_metric['metric'].get('name', None)
            if not pod or not name:
                continue

            times, values = map(list, zip(*container_metric['values']))
            times = list(map(lambda x: datetime.fromtimestamp(x).strftime('%Y-%m-%d %H:%M:%S'), times))
            values = np.array(values, dtype=np.float64)
            if field not in series:
                series[field] = values
                timestamps = times
            else:
                diff = series[field].size - values.size
                # 긴쪽이 더 최신이므로, 짧은 쪽에 0인 값 추가
                if 1 < diff or diff < -1:
                    continue
                if diff == 1:
                    values = np.append(values, 0)
                elif diff == -1:
                    timestamps.append('')  # frequency가 정확히 일치하지않아서 frequency 맞춰주기위해서 나중에 start time으로 재설정
                    series[field] = np.append(series[field], 0)
                series[field] = np.add(series[field], values)  # container가 여러 개인 경우 total로 계산
    timestamps = pd.date_range(start=timestamps[0], periods=len(timestamps), freq=constants.FREQ)
    data = pd.DataFrame(series, index=timestamps)

    return QueryParams(query, data)
