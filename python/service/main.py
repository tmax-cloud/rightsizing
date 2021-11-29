import argparse
from datetime import datetime
import json
import time
import os

from constants import *
from forecast import forecasting
from optimization import extract_resource_usage
from utils import suppress_stdout_stderr

import requests
import numpy as np


CONTAINER_NAME = os.getenv(CONTAINER_NAME_ENV)


def wait_for_server(url):
    print("Waiting for interval server")
    while True:
        try:
            resp = requests.get(url, timeout=1)
            if resp.ok:
                break
        except requests.RequestException:
            time.sleep(10)


parser = argparse.ArgumentParser(description='RightSizing parameter parser')

parser.add_argument('--url', '-u', type=str, required=True, help='Prometheus URL for requests')
parser.add_argument('--manage_server_url', '-server_url', type=str, required=True, help='The manager server url')
parser.add_argument('--optimization', dest="optimization", action='store_true')
parser.add_argument('--no-optimization', dest="optimization", action='store_false')
parser.add_argument('--forecast', dest="forecast", action='store_true')
parser.add_argument('--no-forecast', dest="forecast", action='store_false')
parser.add_argument('--term', '-t', type=str, default="8d", help='query term (default 8d, 8days)')
parser.add_argument('--query', '-q', type=str, required=True)
args = parser.parse_args()

url = args.url
if url.endswith('/'):
    url = url[:-1]
url = url + QUERY_ENDPOINT

server_url = args.manage_server_url
if server_url.endswith('/'):
    server_url = server_url[:-1]


query = {
    'query': f'{args.query}[{args.term}]'
}

r = requests.get(url, query, timeout=300)
if r.status_code != 200:
    raise Exception("request exeception")
request_data = r.json()['data']['result']

times, values = map(list, zip(*request_data[0]['values']))
times = [datetime.fromtimestamp(date) for date in times]
values = np.array(values, dtype=np.float32)


result = dict()
if args.optimization:
    optimization_data = extract_resource_usage(values, 95)
    result['optimization'] = {
        'data': optimization_data
    }

if args.forecast:
    # remove fbprophet log
    with suppress_stdout_stderr():
        result["forecast"] = forecasting(times, values)

wait_for_server(server_url + '/ready')

endpoint = f"{server_url}/{MANAGE_SERVER_ENDPOINT}/{CONTAINER_NAME}"

if result:
    data = json.dumps(result).encode('utf-8')
    r = requests.put(endpoint, data=data, timeout=10)
    if not r.ok:
        raise Exception('request error')
