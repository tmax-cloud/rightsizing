import asyncio

from fields import *

import aiohttp


async def async_get(session, url, query):
    async with session.get(url, params=query) as resp:
        result = await resp.json()
        return result


async def instant_query_prometheus(url: str):
    url = url + '/api/v1/query'

    async with aiohttp.ClientSession() as session:
        tasks = list()
        for field in instant_fields:
            query = {
                'query': field
            }
            tasks.append(asyncio.create_task(async_get(session, url, query)))
        result = await asyncio.gather(*tasks)
    return result


async def query_pod_prometheus(url: str, field: str, nodes: set):
    url = url + '/api/v1/query'

    # field_format = 'avg(rate({field}{{container!="",container!="POD",job="kubelet",namespace="{name}"}}[30d])) ' \
    #                'by (namespace, pod, container) ' \
    #                '* on(namespace, pod, container) group_left() (kube_pod_container_status_running == 1)'
    # field_format = 'avg(rate({field}{{container!="",container!="POD",job="kubelet",node="{node_name}"}}[30d])) ' \
    #                'by (namespace, pod, container) ' \
    #                '* on(namespace, pod, container) group_left() (kube_pod_container_status_running == 1)'
    async with aiohttp.ClientSession() as session:
        tasks = list()
        for node_name in nodes:
            query = {
                'query': field.format(node_name=node_name)
            }
            tasks.append(asyncio.create_task(async_get(session, url, query)))
        result = await asyncio.gather(*tasks)
    return result
