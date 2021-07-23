import warnings

warnings.filterwarnings(action='ignore')

import pandas as pd
from sklearn.ensemble import IsolationForest
from statsmodels.tsa.arima.model import ARIMA

from constants import *


def forecasting(data: pd.DataFrame, value_field: str):
    model = ARIMA(data[value_field], order=(1, 0, 0))
    model_fit = model.fit()

    times = pd.date_range(start=data.index[-1], freq=FREQ, periods=FORECAST_STEP + 1, closed="right")
    forecast = model_fit.predict(start=times[0], end=times[-1])

    return pd.Series(data=forecast, index=times)


def abnormal_detection(data: pd.DataFrame, field: str):
    outliers_fraction = float(.01)

    model = IsolationForest(contamination=outliers_fraction)
    model.fit(data)

    return model.predict(data)
