import argparse
import datetime
import os
import time

from query import *

from apscheduler.schedulers.background import BlockingScheduler
import requests


QUERY_ENDPOINT = '/api/v1/query'
ALERT_ENDPOINT = '/alert'
FORECAST_ENDPOINT = '/forecast'

FUTURE_USAGE_THRESHOLD = os.getenv('usage_threshold', 0.8)
LIMIT_THRESHOLD = os.getenv('limit_threshold', 0.8)
if not isinstance(LIMIT_THRESHOLD, float):
    raise TypeError('"limit_threshold" environments must be float')


sched = BlockingScheduler()


def prometheus_query(url: str, query: str):
    r = requests.get(url, {'query': query})
    if r.status_code != 200:
        raise Exception("request exeception")
    data = r.json()['data']['result']
    if not data:
        return None
    return data[0]['value'][0]


def tracking(url: str, namespace: str, pod_name: str, server_url: str):
    print('tracking at', datetime.datetime.now())
    results = list()

    for limit_field, request_field, usage_field in zip(limit_fields, request_fields, resource_usage_field):
        limit_result = prometheus_query(url, limit_field.format(namespace=namespace, pod_name=pod_name))
        usage_result = prometheus_query(url, usage_field.format(namespace=namespace, pod_name=pod_name))
        results.append((limit_result, usage_result))

    for resource, result in zip(resources_order, results):
        limit, usage = result

        r = requests.get(server_url + FORECAST_ENDPOINT)
        if r.status_code == 200:
            future_usage = r.json()
            if future_usage * FUTURE_USAGE_THRESHOLD < usage:
                requests.put(server_url, {resource: 'warning'})

        if limit is None:
            print("No resource limits registered. Please register resource limit")
        else:
            if limit * LIMIT_THRESHOLD < usage:
                requests.put(server_url, {resource: 'dangerous'})


parser = argparse.ArgumentParser(description='RightSizing parameter parser')

parser.add_argument('--name', '-n', type=str, required=True, help='The pod name for rightsizing')
parser.add_argument('--prometheus_url', '-url', type=str, required=True, help='Prometheus URL for requests')
parser.add_argument('--namespace', '-ns', type=str, default="", help='Namespace to identify pod (result must be unique)')
parser.add_argument('--manage_server_url', '-server_url', type=str, required=True, help='The manager server url')
parser.add_argument('--interval', '-i', type=int, default=600, help='interval time (second) for tracking (default 600 seconds)')


args = parser.parse_args()

prometheus_url = args.prometheus_url
if prometheus_url.endswith('/'):
    prometheus_url = prometheus_url[:-1]
if not prometheus_url.endswith(QUERY_ENDPOINT):
    prometheus_url = prometheus_url + QUERY_ENDPOINT

server_url = args.manage_server_url
if server_url.endswith('/'):
    server_url = prometheus_url[:-1]

name = args.name
namespace = args.namespace

interval = args.interval


for i in range(10):
    try:
        r = requests.get(server_url + FORECAST_ENDPOINT + '/ready', timeout=3)
        if r.ok:
            break
    except:
        print("waiting for server ready")
        time.sleep(10)

sched.add_job(tracking, 'interval', seconds=interval, args=[prometheus_url, namespace, name, server_url])
print("scheduler start ...")
sched.start()
