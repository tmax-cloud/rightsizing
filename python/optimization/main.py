import argparse
from datetime import datetime
import json
import time

from util import suppress_stdout_stderr

import requests
import numpy as np

MANAGER_ENDPOINT = '/optimization'
QUERY_ENDPOINT = '/api/v1/query'

FIELDS = [
    'pod:container_cpu_usage:sum',
    'pod:container_memory_usage_bytes:sum'
]

FIELD_NAMES = [
    "cpu",
    "memory"
]

PERCENTILES = [
    99, 98, 95, 90, 80
]
MARGIN = 0.2


def quantiles(data: np.ndarray, quantile: int) -> np.ndarray:
    q = PERCENTILES[quantile]
    m = np.percentile(data, q)
    return m


def extract_resource_usage(data: np.ndarray, sensitivity):
    percentiles = quantiles(data, sensitivity)
    return percentiles * (1 + MARGIN)


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
sensitivity = args.sensitivity

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
        r = requests.get(url, query)
        if r.status_code != 200:
            raise Exception("request exeception")
        request_data = r.json()['data']['result']

        times, values = map(list, zip(*request_data[0]['values']))
        times = [datetime.fromtimestamp(date) for date in times]
        values = np.array(values, dtype=np.float32)

        result[field_name] = extract_resource_usage(values, sensitivity)

data = json.dumps(result).encode('utf-8')

for i in range(10):
    try:
        r_put = requests.put(server_url, data=data, timeout=3)
        if r_put.ok:
            break
    except:
        print("waiting for server ready")
        time.sleep(10)
