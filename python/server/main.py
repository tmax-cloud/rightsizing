import datetime
from typing import Dict, List, Optional

from fastapi import FastAPI, Response, Request, status
from pydantic import BaseModel


class ForecastData(BaseModel):
    ds: List[datetime.datetime] = []
    yhat: List[float] = []
    yhat_upper: List[float] = []
    yhat_lower: List[float] = []


class ForecastSummary(BaseModel):
    ds: datetime.datetime
    yhat: float
    yhat_upper: float
    yhat_lower: float


class ForecastRequest(BaseModel):
    data: ForecastData


class OptimizationRequest(BaseModel):
    data: float


class Item(BaseModel):
    forecast: Optional[ForecastData] = None
    optimization: Optional[OptimizationRequest] = None


class QueryData(BaseModel):
    forecast: Optional[ForecastSummary] = None
    optimization: Optional[OptimizationRequest] = None


app = FastAPI()

app.data = dict()


@app.get("/ready")
def ready():
    return True


@app.put("/queries/{name}")
def receive_query_data(name: str, data: Item):
    app.data[name] = data


@app.get("/queries/{name}/forecast", response_model=ForecastRequest, status_code=200)
def send_forecast_data(name: str):
    if name not in app.data or app.data[name].forecast is None:
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    return {
        'data': app.data[name].forecast
    }


@app.get("/queries/{name}/optimization", response_model=OptimizationRequest, status_code=200)
def send_optimization_data(name: str):
    if name not in app.data or app.data[name].optimization is None:
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    return app.data[name].optimization


@app.get("/queries/{name}", response_model=QueryData, status_code=200)
def send_query_data(name: str):
    if name not in app.data:
        return Response(status_code=status.HTTP_204_NO_CONTENT)

    data = QueryData()

    query_result = app.data[name]
    if query_result.forecast:
        forecast = query_result.forecast
        data.forecast = ForecastSummary(
            ds=forecast.ds[-1],
            yhat=forecast.yhat[-1],
            yhat_upper=forecast.yhat_upper[-1],
            yhat_lower=forecast.yhat_lower[-1])
    if query_result.optimization:
        data.optimization = query_result.optimization
    return data
