import argparse
from datetime import datetime
import json
import time

from constants import *
from forecast import forecasting
from optimization import extract_resource_usage
from utils import suppress_stdout_stderr

import requests
import numpy as np


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
parser.add_argument('--term', '-t', type=int, default=8, help='query term (days) (7 mean 7 days)')
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
    'query': f'{args.query}[{args.term}d]'
}

r = requests.get(url, query, timeout=300)
if r.status_code != 200:
    raise Exception("request exeception")
request_data = r.json()['data']['result']

times, values = map(list, zip(*request_data[0]['values']))
times = [datetime.fromtimestamp(date) for date in times]
values = np.array(values, dtype=np.float32)


results = dict()
if args.optimization:
    optimization_result = extract_resource_usage(values, 95)
    result = {'data': optimization_result}
    results["optimization"] = json.dumps(result).encode('utf-8')

if args.forecast:
    with suppress_stdout_stderr():
        forecast_result = forecasting(times, values)
        result = {'data': forecast_result}
        results["forecast"] = json.dumps(result).encode('utf-8')

wait_for_server(server_url + '/ready')

for service_name, data in results.items():
    endpoint = f"{server_url}/{service_name}"
    r = requests.put(endpoint, data=data, timeout=10)
    if not r.ok:
        raise Exception('request error')
