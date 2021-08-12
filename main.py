from fastapi import FastAPI, Depends
from fastapi.middleware.wsgi import WSGIMiddleware
from fastapi_cache import FastAPICache
from fastapi_cache.backends.inmemory import InMemoryBackend
from fastapi_cache.coder import PickleCoder
from fastapi_cache.decorator import cache
from starlette.status import HTTP_201_CREATED, HTTP_204_NO_CONTENT

from models import *
from plotly_dash import create_dash_app
from query import query_and_analyze_pod
from utils import redis_key_builder
import constants
import ML

logger = logging.getLogger(__name__)

app = FastAPI()

# dash_app = create_dash_app('/dash/')
# app.mount('/', WSGIMiddleware(dash_app.server))


@app.get("/health")
async def health():
    return {"health": "ok"}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@app.post('/forecast')
def forecast_usage(query: QueryParams = Depends(query_and_analyze_pod, use_cache=True)):
    if query is None:
        return {}
    data = query.data

    cpu_df = data[constants.CPU_FIELD].to_frame().reset_index().rename(
        columns={"index": "ds", constants.CPU_FIELD: "y"})
    _, cpu_forecast = ML.forecasting(cpu_df)

    memory_df = data[constants.MEMORY_FIELD].to_frame().reset_index().rename(
        columns={"index": "ds", constants.MEMORY_FIELD: "y"})
    _, memory_forecast = ML.forecasting(memory_df)

    return {'CPU': cpu_forecast, "Memory": memory_forecast}


@cache(expire=300, coder=PickleCoder, key_builder=redis_key_builder)
@app.post('/abnormal_detect')
def abnormal_detect(query: QueryParams = Depends(query_and_analyze_pod, use_cache=True)):
    if query is None:
        return {}
    scaled_data = query.scaled_data

    cpu_abnormal = ML.abnormal_detection(scaled_data)
    memory_abnormal = ML.abnormal_detection(scaled_data)

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

    _, cpu_forecast = ML.forecasting(data[constants.CPU_FIELD])
    cpu_forecast = query.inverse_transform(cpu_forecast, constants.CPU_FIELD)
    _, memory_forecast = ML.forecasting(data[constants.MEMORY_FIELD])
    memory_forecast = query.inverse_transform(memory_forecast, constants.MEMORY_FIELD)

    cpu_abnormal = ML.abnormal_detection(data)
    memory_abnormal = ML.abnormal_detection(data)

    return {
        "CPU Forecast": cpu_forecast,
        "CPU Abnormal": cpu_abnormal.tolist(),
        "Memory Forecast": memory_forecast,
        "Memory abnormal": memory_abnormal.tolist()
    }


@app.on_event("startup")
async def startup():
    FastAPICache.init(InMemoryBackend(), prefix="fastapi-cache")

    logger.info("Successfully initialize internal cache")