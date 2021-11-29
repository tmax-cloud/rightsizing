from datetime import datetime, timedelta
from typing import List

from fbprophet import Prophet
import numpy as np
import pandas as pd


def forecasting(ds: List, data: np.ndarray):
    df = pd.DataFrame({'ds': ds, 'origin': data})
    df['y'] = df['origin'].ewm(halflife="12 hours", times=df['ds']).mean()

    m = Prophet(interval_width=0.1)
    m.fit(df)
    future = m.make_future_dataframe(periods=1440, freq='min')
    forecast = m.predict(future)
    forecast['ds'] = forecast['ds'].astype(str)

    result = dict()

    now = datetime.now()
    end_time = now + timedelta(hours=6)

    d = forecast[['ds', 'yhat', 'yhat_lower', 'yhat_upper']]
    d = d.query(f"ds >= '{now:%Y-%m-%d %H:%M}' and ds <= '{end_time:%Y-%m-%d %H:%M}'")
    for key, series in d.items():
        result[key] = list(series)
    return result
