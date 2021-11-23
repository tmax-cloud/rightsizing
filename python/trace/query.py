interval = "5m"

limit_fields = [
    # metric fields
    'kube_pod_container_resource_limits_cpu_cores{{container!="POD",namespace="{namespace}",pod="{pod_name}"}}',
    'kube_pod_container_resource_limits_memory_bytes{{container!="POD",namespace="{namespace}",pod="{pod_name}"}}',
]

request_fields = [
    'kube_pod_container_resource_requests_cpu_cores{{container!="POD",namespace="{namespace}",pod="{pod_name}"}}',
    'kube_pod_container_resource_requests_memory_bytes{{container!="POD",namespace="{namespace}",pod="{pod_name}"}}'
]

resource_usage_field = [
    f'avg(rate(container_cpu_usage_seconds_total{{{{container!="",container!="POD",job="kubelet",namespace="{{namespace}}",pod="{{pod_name}}"}}}}[{interval}]))',
    f'avg_over_time(container_memory_usage_bytes{{{{container!="",container!="POD",job="kubelet",namespace="{{namespace}}",pod="{{pod_name}}"}}}}[{interval}])'
]

# %%Pay attention%%
# The order of this field means query order
resources_order = [
    'CPU',
    'Memory'
]