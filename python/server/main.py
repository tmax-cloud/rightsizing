import datetime
from typing import Dict, List, Optional

from fastapi import FastAPI, Response, Request, status
from pydantic import BaseModel


class ForecastData(BaseModel):
    ds: List[datetime.datetime] = []
    yhat: List[float] = []
    yhat_upper: List[float] = []
    yhat_lower: List[float] = []

class ForecastSummaryData(BaseModel):
    ds: datetime.datetime
    yhat: float
    yhat_upper: float
    yhat_lower: float

class ForecastRequest(BaseModel):
    cpu: ForecastData
    memory: ForecastData

class OptimizationRequest(BaseModel):
    cpu: float
    memory: float


app = FastAPI()

app.is_warning = False
app.forecast = None
app.optimization = None


@app.get("/ready")
def ready():
    return True


@app.put("/alert")
def alert_signal():
    app.is_warning = True


@app.get("/alert")
async def send_is_alert():
    return app.is_warning


@app.get("/forecast", response_model=ForecastRequest, status_code=200)
def send_forecast_data():
    if not app.forecast:
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    return app.forecast

@app.get("/forecast/summary", response_model=ForecastSummaryData, status_code=200)
def send_forecast_summary_data():
    if not app.forecast or len(app.forecast.yhat) < 1:
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    return {
        'ds': app.forecast.ds[-1],
        'yhat': app.forecast.yhat[-1],
        'yhat_upper': app.forecast.yhat_upper[-1],
        'yhat_lower': app.forecast.yhat_lower[-1]
    }



@app.put("/forecast")
async def receive_forecast_data(request: ForecastRequest):
    app.forecast = request


@app.get("/forecast/ready")
async def forecast_ready():
    if not app.forecast:
        return False
    return True


@app.put("/optimization")
def receive_optimization_data(request: OptimizationRequest):
    app.optimization = request


@app.get("/optimization", response_model=OptimizationRequest, status_code=200)
def send_optimization_data():
    if not app.optimization:
        return Response(status_code=status.HTTP_204_NO_CONTENT)
    return app.optimization


@app.get("/optimization/ready")
async def optimization_ready():
    if not app.optimization:
        return False
    return True
