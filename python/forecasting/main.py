import argparse
from datetime import datetime, timedelta
import time
from typing import List
import json

from util import suppress_stdout_stderr

from fbprophet import Prophet
import requests
import numpy as np
import pandas as pd


MANAGER_ENDPOINT = '/forecast'
QUERY_ENDPOINT = '/api/v1/query'

FIELDS = [
    'pod:container_cpu_usage:sum',
    'pod:container_memory_usage_bytes:sum'
]

FIELD_NAMES = [
    "cpu",
    "memory"
]


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


parser = argparse.ArgumentParser(description='RightSizing parameter parser')

parser.add_argument('--name', '-n', type=str, required=True, help='The pod name for rightsizing')
parser.add_argument('--prometheus_url', '-url', type=str, required=True, help='Prometheus URL for requests')
parser.add_argument('--namespace', '-ns', type=str, default="", help='Namespace to identify pod (result must be unique)')
parser.add_argument('--sensitivity', '-s', type=int, default=0, help='The sensitivity for resource (0-2)')
parser.add_argument('--manage_server_url', '-server_url', type=str, required=True, help='The manager server url')

args = parser.parse_args()

url = args.prometheus_url
server_url = args.manage_server_url
name = args.name
namespace = args.namespace

if url.endswith('/'):
    url = url[:-1]
url = url + QUERY_ENDPOINT

if server_url.endswith('/'):
    server_url = server_url[:-1]
server_url = f'{server_url}{MANAGER_ENDPOINT}'

query_fmt = f'{{pod="{name}",namespace="{namespace}"}}'


with suppress_stdout_stderr():
    result = dict()
    for field_name, field in zip(FIELD_NAMES, FIELDS):
        query = {
            'query': f'{field}{query_fmt}[7d]'
        }
        r = requests.get(url, query, timeout=300)
        if r.status_code != 200:
            raise Exception("request exeception")
        request_data = r.json()['data']['result']

        times, values = map(list, zip(*request_data[0]['values']))
        times = [datetime.fromtimestamp(date) for date in times]
        values = np.array(values, dtype=np.float32)

        result[field_name] = forecasting(times, values)

data = json.dumps(result).encode('utf-8')

for i in range(10):
    try:
        r_put = requests.put(server_url, data=data, timeout=10)
        if r_put.ok:
            break
    except:
        print("waiting for server ready")
        time.sleep(10)

