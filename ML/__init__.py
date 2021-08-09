import warnings

import pandas as pd
from sklearn.ensemble import IsolationForest
from statsmodels.tsa.arima_model import ARIMA

from constants import *

warnings.filterwarnings(action='ignore')


def forecasting(data: pd.DataFrame, value_field: str):
    model = ARIMA(data, order=(1, 0, 0))
    model_fit = model.fit()

    times = pd.date_range(start=data.index[-1], freq=FREQ, periods=FORECAST_STEP + 1, closed="right")
    forecast = model_fit.predict(start=times[0], end=times[-1])

    return pd.Series(data=forecast, index=times)


def abnormal_detection(data: pd.DataFrame, field: str):
    outliers_fraction = float(.01)

    model = IsolationForest(contamination=outliers_fraction)
    model.fit(data)

    return model.predict(data)
