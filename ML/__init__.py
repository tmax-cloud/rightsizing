import warnings

from fbprophet import Prophet
import pandas as pd
from sklearn.ensemble import IsolationForest

import constants

warnings.filterwarnings(action='ignore')


def forecasting(data: pd.DataFrame):
    m = Prophet()
    m.fit(data)

    future = m.make_future_dataframe(periods=constants.FORECAST_STEP, freq=constants.FREQ)
    forecast = m.predict(future)

    return m, forecast


def abnormal_detection(data: pd.DataFrame):
    outliers_fraction = float(.01)

    model = IsolationForest(contamination=outliers_fraction)
    model.fit(data)

    return model.predict(data)
