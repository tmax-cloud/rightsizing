from fastapi import APIRouter, Depends

from fastapi_cache import FastAPICache
from fastapi_cache.backends.inmemory import InMemoryBackend
from fastapi_cache.coder import PickleCoder
from fastapi_cache.decorator import cache
from starlette.status import HTTP_201_CREATED

import ML
import constants
from .models import *
from .query import query_and_analyze_pod
from .utils import redis_key_builder

router = APIRouter()


async def common_query(q: QueryParams = Depends(query_and_analyze_pod)):
    return q


@router.get("/health")
async def health():
    return {"health": "ok"}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@router.post('/forecast')
def forecast_usage(query: QueryParams = Depends(common_query, use_cache=True)):
    scaled_data = query.scaled_data

    cpu_forecast = ML.forecasting(scaled_data, constants.CPU_FIELD)
    cpu_forecast = query.inverse_transform(cpu_forecast, constants.CPU_FIELD)

    memory_forecast = ML.forecasting(scaled_data, constants.MEMORY_FIELD)
    memory_forecast = query.inverse_transform(memory_forecast, constants.MEMORY_FIELD)

    return {'CPU': cpu_forecast, "Memory": memory_forecast}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@router.post('/abnormal_detect')
def abnormal_detect(query: QueryParams = Depends(common_query, use_cache=True)):
    scaled_data = query.scaled_data

    cpu_abnormal = ML.abnormal_detection(scaled_data, constants.CPU_FIELD)
    memory_abnormal = ML.abnormal_detection(scaled_data, constants.MEMORY_FIELD)

    return {'CPU': cpu_abnormal.tolist(), "Memory": memory_abnormal.tolist()}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@router.post('/analysis/', status_code=HTTP_201_CREATED)
def create_analysis_task(query: QueryParams = Depends(common_query, use_cache=True)):
    """
    Create analysis report for kubernetes object (pod).
    This will let the API user create an analysis report.

    - **url**: The url of prometheus server (ex. https://console.tmaxcloud.com/api/grafana/api/datasources/proxy/1)
    - **namespace**: The namespace of kubernetes object to identify
    - **name**: The name of kubernetes object
    \f
    :param query: Prometheus query result DataFrame.
    :param background_tasks: BackgroundTasks
    """

    data = query.scaled_data

    cpu_forecast = ML.forecasting(data, constants.CPU_FIELD)
    cpu_forecast = query.inverse_transform(cpu_forecast, constants.CPU_FIELD)
    memory_forecast = ML.forecasting(data, constants.MEMORY_FIELD)
    memory_forecast = query.inverse_transform(memory_forecast, constants.MEMORY_FIELD)

    cpu_abnormal = ML.abnormal_detection(data, constants.CPU_FIELD)
    memory_abnormal = ML.abnormal_detection(data, constants.MEMORY_FIELD)

    return {
        "CPU Forecast": cpu_forecast,
        "CPU Abnormal": cpu_abnormal.tolist(),
        "Memory Forecast": memory_forecast,
        "Memory abnormal": memory_abnormal.tolist()
    }


@router.on_event("startup")
async def startup():
    # redis = await aioredis.create_redis_pool("redis://localhost", encoding="utf8")
    # FastAPICache.init(RedisBackend(redis), prefix="fastapi-cache")
    FastAPICache.init(InMemoryBackend(), prefix="fastapi-cache")
