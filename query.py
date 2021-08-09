from datetime import datetime
import logging

from fastapi import Depends
from fastapi_cache.decorator import cache
from fastapi_cache.coder import PickleCoder
import pandas as pd
import numpy as np

import constants
from db import DBConnection, get_db
from models import AnalysisQuery, QueryParams
from utils import redis_key_builder

logger = logging.getLogger(__name__)

_query_fmt = '''
select mean(value) from {field} 
WHERE pod = '{name}' and namespace = '{namespace}' and time > '{start}' - {duration}
group by container, time(30s) fill(linear);'''


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
async def query_and_analyze_pod(
    db: DBConnection = Depends(get_db),
    *,
    query: AnalysisQuery):
    """Query and analyze kubernetes pod.

    :param namespace: The namespace of kubernetes object to identify
    :param name: The name of kubernetes object
    :return: pd.DataFrame
    """

    namespace = query.namespace
    name = query.name

    now = str(datetime.now())
    # pod query string
    pod_fields = ','.join([field for field in constants.FIELDS])
    pod_query = _query_fmt.format(field=pod_fields, name=name, namespace=namespace,
                                  start=now, duration=constants.QUERY_TERM)

    result = db.query(pod_query)
    if result.empty:
        return None

    total_df = pd.DataFrame()
    for record in result:
        for key, data in record.items():
            field, _ = key.split(',')
            total_df[field] = np.array(data['mean'], dtype=np.float32)
    total_df = total_df.fillna(method='ffill')
    return QueryParams(query=query, data=total_df)
