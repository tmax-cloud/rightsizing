from fastapi import FastAPI, Depends, Request
from fastapi_cache import FastAPICache
from fastapi_cache.backends.inmemory import InMemoryBackend
from fastapi_cache.coder import PickleCoder
from fastapi_cache.decorator import cache
from starlette.status import HTTP_201_CREATED, HTTP_204_NO_CONTENT

from db import DBConnection
from models import *
from query import query_and_analyze_pod
from utils import redis_key_builder
import constants
import ML

logger = logging.getLogger(__name__)

app = FastAPI()


@app.get("/health")
async def health():
    return {"health": "ok"}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@app.post('/forecast')
def forecast_usage(query: QueryParams = Depends(query_and_analyze_pod, use_cache=True)):
    if query is None:
        return {}
    scaled_data = query.scaled_data

    cpu_forecast = ML.forecasting(scaled_data, constants.CPU_FIELD)
    cpu_forecast = query.inverse_transform(cpu_forecast, constants.CPU_FIELD)

    memory_forecast = ML.forecasting(scaled_data, constants.MEMORY_FIELD)
    memory_forecast = query.inverse_transform(memory_forecast, constants.MEMORY_FIELD)

    return {'CPU': cpu_forecast, "Memory": memory_forecast}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@app.post('/abnormal_detect')
def abnormal_detect(query: QueryParams = Depends(query_and_analyze_pod, use_cache=True)):
    if query is None:
        return {}
    scaled_data = query.scaled_data

    cpu_abnormal = ML.abnormal_detection(scaled_data, constants.CPU_FIELD)
    memory_abnormal = ML.abnormal_detection(scaled_data, constants.MEMORY_FIELD)

    return {'CPU': cpu_abnormal.tolist(), "Memory": memory_abnormal.tolist()}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@app.post('/analysis/', status_code=HTTP_201_CREATED)
def create_analysis_task(query: QueryParams = Depends(query_and_analyze_pod, use_cache=True)):
    """
    Create analysis report for kubernetes object (pod).
    This will let the API user create an analysis report.

    - **url**: The url of prometheus server (ex. https://console.tmaxcloud.com/api/grafana/api/datasources/proxy/1)
    - **namespace**: The namespace of kubernetes object to identify
    - **name**: The name of kubernetes object
    \f
    :param query: Prometheus query result DataFrame.
    """
    if query is None:
        return {}

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


@app.on_event("startup")
async def startup():
    FastAPICache.init(InMemoryBackend(), prefix="fastapi-cache")

    influxdb = DBConnection()
    app.state.db = influxdb

    logger.info("Initializing InfluxDB connection")


@app.on_event("shutdown")
async def shutdown():
    app.state.db.close()
    logger.info("Close InfluxDB connection")
