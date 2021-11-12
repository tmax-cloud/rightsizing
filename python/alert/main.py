import asyncio
import argparse
from datetime import datetime
import json

from fields import *
from query import instant_query_prometheus, query_pod_prometheus

import requests


parser = argparse.ArgumentParser(description='RightSizing parameter parser')
parser.add_argument('--url', '-u', type=str, required=True, help='Prometheus URL for requests')
args = parser.parse_args()

url = args.url


TOKEN = 'xoxb-2552037846225-2539421227523-se1i73iQQupH5l01X7X3UQkN'
channel_id = 'C02FSBYFP61'
# url = "https://console.tmaxcloud.com/api/grafana/api/datasources/proxy/1"
# url = "http://192.168.9.242:9090"
upper_limit = 0.8
lower_limit = 0.1
storage_limit = 0.5

print(datetime.now(), "start query information of all pod in kubernetes to prometheus")

instant_query_result = asyncio.run(instant_query_prometheus(url))

# initial pod resource usage dictionary in advance
pods = dict()
nodes = set()

pod_counter = 0
for entry in instant_query_result[0]['data']['result']:
    namespace, pod_name = entry['metric']['namespace'], entry['metric']['pod']
    if namespace not in pods:
        pods[namespace] = dict()
    pods[namespace][pod_name] = dict(container=dict())
    pod_counter += 1
# initial volume dictionary for mapping volume to pod
volumes = dict()
# mapping persistent-volume to pod (pod and pvc have 1:1 relationship)
for entry in instant_query_result[1]['data']['result']:
    namespace = entry['metric']['namespace']
    pvc_name = entry['metric']['persistentvolumeclaim']
    if namespace not in volumes:
        volumes[namespace] = dict()
    # mapping relationship in each dictionary
    volumes[namespace][pvc_name] = dict()
# mapping container to pod (pod and container are 1:n relationship)
for entry in instant_query_result[2]['data']['result']:
    namespace = entry['metric']['namespace']
    pod_name = entry['metric']['pod']
    container_name = entry['metric']['container']
    pods[namespace][pod_name]['container'][container_name] = dict(request=dict(), limit=dict(), usage=dict())
# get nodes in kubernetes
for entry in instant_query_result[3]['data']['result']:
    node_name = entry['metric']['node']
    nodes.add(node_name)
# mapping each resource metric values to container
for i, metric in enumerate(instant_query_result[4:7]):
    data = metric['data']['result']
    for entry in data:
        namespace = entry['metric']['namespace']
        pod_name = entry['metric']['pod']
        container_name = entry['metric']['container']
        metric_type, metric_name = resource_explain_field[i]
        value = entry['value'][1]
        pods[namespace][pod_name]['container'][container_name][metric_type][metric_name] = float(value)
# mapping storage metric values
for i, metric in enumerate(instant_query_result[-2:]):
    data = metric['data']['result']
    for entry in data:
        pvc_name = entry['metric']['persistentvolumeclaim']
        namespace = entry['metric']['exported_namespace']
        metric_type = storage_explain_field[i]
        value = entry['value'][1]
        volumes[namespace][pvc_name][metric_type] = float(value)
print(datetime.now(), "instant_request process done")

print(datetime.now(), "pod resource usage query")
cpu_usage_result = asyncio.run(query_pod_prometheus(url, resource_usage_field[0], nodes))
memory_usage_result = asyncio.run(query_pod_prometheus(url, resource_usage_field[1], nodes))

print(datetime.now(), "pod resource usage query process")
# CPU
for metric in cpu_usage_result:
    data = metric['data']['result']
    for entry in data:
        namespace = entry['metric']['namespace']
        pod_name = entry['metric']['pod']
        container_name = entry['metric']['container']
        value = entry['value'][1]
        pods[namespace][pod_name]['container'][container_name]['usage']['cpu_cores'] = float(value)
# memory
for metric in memory_usage_result:
    data = metric['data']['result']
    for entry in data:
        namespace = entry['metric']['namespace']
        pod_name = entry['metric']['pod']
        container_name = entry['metric']['container']
        value = entry['value'][1]
        pods[namespace][pod_name]['container'][container_name]['usage']['memory_bytes'] = float(value)
print(datetime.now(), "pod resource usage query process done")

alert_pods = set()
alert_cpu_pods = set()
alert_memory_pods = set()

alert_cpu_pod_list = list()
alert_memory_pod_list = list()
for namespace, data in pods.items():
    # pod
    for pod_name, pod_data in data.items():
        # container
        for container_name, container_data in pod_data['container'].items():
            if 'cpu_cores' in container_data['limit'] and 'cpu_cores' in container_data['usage']:
                cpu_limit = container_data['limit']['cpu_cores']
                cpu_usage = container_data['usage']['cpu_cores']
                if cpu_usage / cpu_limit < lower_limit:
                    alert_cpu_pod_list.append((namespace, pod_name, container_name, cpu_usage / cpu_limit))
                    alert_cpu_pods.add(f'{namespace}_{pod_name}')
            if 'memory_bytes' in container_data['limit'] and 'memory_bytes' in container_data['usage']:
                memory_limit = container_data['limit']['memory_bytes']
                memory_usage = container_data['usage']['memory_bytes']
                if memory_usage / memory_limit < lower_limit:
                    alert_memory_pod_list.append((namespace, pod_name, container_name, memory_usage / memory_limit))
                    alert_memory_pods.add(f'{namespace}_{pod_name}')
alert_volume_list = list()
for namespace, data in volumes.items():
    for pvc_name, pvc_data in data.items():
        if 'capacity' in pvc_data and 'used' in pvc_data:
            volume_capacity = pvc_data['capacity']
            volume_used = pvc_data['used']
            if volume_used / volume_capacity > 0.5:
                alert_volume_list.append(pvc_name)
print(datetime.now(), "done")

alert_pods = alert_cpu_pods.union(alert_memory_pods)
if pod_counter < 1:
    alert_ratio = 0.
else:
    alert_ratio = len(alert_pods) / pod_counter

alert_cpu_pod_list = sorted(alert_cpu_pod_list, key=lambda pod: pod[3])
alert_memory_pod_list = sorted(alert_memory_pod_list, key=lambda pod: pod[3])

msg = [
    {
        "type": "divider"
    },
    {
        "type": "header",
        "text": {
            "text": ":warning: RightSizing Alert Service",
            "type": "plain_text"
        }
    },
    {
        "type": "context",
        "elements": [
            {
                "text": f"현재 prometheus 주소는 {url} 입니다.\n"
                        # f"⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼\n"
                        f"실행 중인 pod 개수는 {pod_counter}개이며 이 중 {len(alert_pods)}가 과할당 또는 저할당 되었습니다 "
                        f"(약 {alert_ratio * 100:.1f}%).\n",
                "type": "mrkdwn"
            }
        ]
    },
    {
        "type": "section",
        "text": {
            "text": "*⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼ _CPU_ ⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼*",
            "type": "mrkdwn"
        }
    },
    {
        "type": "context",
        "elements": [
            {
                "text": f'총 {pod_counter} 중 {len(alert_cpu_pods)}개 '
                        f'(약 {len(alert_cpu_pods) / pod_counter * 100:.1f}%)가 {lower_limit * 100:.0f}% 이하 사용중입니다.\n',
                "type": "mrkdwn"
            }
        ]
    },
    {
        "type": "section",
        "text": {
            "text": '*_사용량 하위 5개_*\n'
                    '> ' + "\n> ".join([", ".join(pod[:3]) for pod in alert_cpu_pod_list][:5]),
            "type": "mrkdwn"
        },
    },
    {
        "type": "section",
        "text": {
            "text": "*⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼ _Memory_ ⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼*",
            "type": "mrkdwn"
        }
    },
    {
        "type": "context",
        "elements": [
            {
                "text": f'총 {pod_counter} 중 {len(alert_memory_pods)}개 '
                        f'(약 {len(alert_memory_pods) / pod_counter * 100:.1f}%)가 {lower_limit * 100:.0f}% 이하 사용중입니다.\n',
                "type": "mrkdwn"
            }
        ]
    },
    {
        "type": "section",
        "text": {
            "text": '*_사용량 하위 5개_*\n'
                    '> ' + "\n> ".join([", ".join(pod[:3]) for pod in alert_memory_pod_list][:5]),
            "type": "mrkdwn"
        },
    },
    {
        "type": "section",
        "text": {
            "text": "*⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼ _Storage_ ⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼⎼*",
            "type": "mrkdwn"
        }
    },
    {
        "type": "context",
        "elements": [
            {
                "text": f'총 {len(volumes.keys())} 중 {len(alert_volume_list)}개 '
                        f'(약 {len(alert_volume_list) / len(volumes.keys()) * 100:.1f}%)가 '
                        f'{storage_limit * 100:.0f}% 이상 사용 중입니다.\n',
                "type": "mrkdwn"
            }
        ]
    },
    {
        "type": "section",
        "text": {
            "text": '> ' + "\n> ".join(alert_volume_list[:5]),
            "type": "mrkdwn"
        },
    }
]

data = {
    'Content-Type': 'application/json; charset=utf-8',
    'token': TOKEN,
    'channel': channel_id,
    'blocks': json.dumps(msg),
    'reply_broadcast': 'True',
}
print(msg)

URL = "https://slack.com/api/chat.postMessage"
res = requests.post(URL, data=data)