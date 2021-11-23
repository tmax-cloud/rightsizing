request_term = "30d"

instant_fields = [
    # information field
    'kube_pod_status_ready{condition="true"}',
    'kube_persistentvolumeclaim_info',
    'kube_pod_container_status_running',
    # node information
    'kube_node_info',
    # metric fields
    ('kube_pod_container_resource_limits_cpu_cores{container!="POD"} * on(namespace, pod) group_left() '
     '(kube_pod_status_phase{phase="Running"} == 1)'),
    ('kube_pod_container_resource_limits_memory_bytes{container!="POD"} * on(namespace, pod) group_left() '
     '(kube_pod_status_phase{phase="Running"} == 1)'),
    ('kube_pod_container_resource_requests_cpu_cores{container!="POD"} * on(namespace, pod) group_left() '
     '(kube_pod_status_phase{phase="Running"} == 1)'),
    ('kube_pod_container_resource_requests_memory_bytes{container!="POD"} * on(namespace, pod) group_left() '
     '(kube_pod_status_phase{phase="Running"} == 1)'),
    # storage field
    'kubelet_volume_stats_capacity_bytes',
    'kubelet_volume_stats_used_bytes'
]

resource_usage_field = [
    ('avg(rate(container_cpu_usage_seconds_total'
     f'{{{{container!="",container!="POD",job="kubelet",node="{{node_name}}"}}}}[{request_term}])) '
     'by (namespace, pod, container) * on(namespace, pod, container) '
     'group_left() (kube_pod_container_status_running == 1)'),
    ('avg_over_time(container_memory_usage_bytes'
     f'{{{{container!="",container!="POD",job="kubelet",node="{{node_name}}"}}}}[{request_term}]) '
     '* on (namespace, pod, container) group_left() (kube_pod_container_status_running == 1)')
]

resource_explain_field = [
    ('limit', 'cpu_cores'),
    ('limit', 'memory_bytes'),
    ('request', 'cpu_cores'),
    ('request', 'memory_bytes'),
]

storage_explain_field = [
    'capacity',
    'used'
]