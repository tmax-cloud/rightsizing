CELERY_BROKER_URL = "CELERY_BROKER_URL"
CELERY_BACKEND_URL = "CELERY_BACKED_URL"

QUERY_ENDPOINT = '/api/v1/query' # Prometheus query endpoint
QUERY_TERM = "7d" # Prometheus query term
# 필터링할 컨테이너 리스트
PAUSE_CONTAINER = "POD"  # Pause container name in Prometheus.

# Query fields to Prometheus
FIELDS = [
    "container_cpu_cfs_throttled_seconds_total",
    "container_memory_cache",
    "container_memory_max_usage_bytes",
    "container_memory_rss",
    "container_memory_swap",
    "container_memory_usage_bytes",
    "container_memory_working_set_bytes",
    "container_fs_io_time_seconds_total",
    "container_fs_io_time_weighted_seconds_total",
]
POD_FIELD = [
    "pod:container_cpu_usage:sum"
    "pod:container_memory_usage_bytes:sum"
]

CPU_FIELD = "container_cpu_cfs_throttled_seconds_total"
MEMORY_FIELD = "container_memory_working_set_bytes"
# Forecasting frequency
FREQ = "30S"
FORECAST_STEP = 1000

# 'sum(rate(container_cpu_usage_seconds_total{name=~".+",namespace=~"$namespace"}[5m])) by (namespace) * 100'
# 'sum by(namespace) (container_memory_working_set_bytes{namespace="$namespace",container="",pod!="",mode!="idle"})'
# 'sum(pod:container_fs_usage_bytes:sum{namespace="$namespace"}) by (namespace)'
# 'sum(rate(container_network_receive_bytes_total{container="POD",pod!="",namespace="$namespace"}[5m]))'
# 'sum(rate(container_network_transmit_bytes_total{container="POD",pod!="",namespace="$namespace"}[5m]))'
# 'count(kube_pod_info{namespace="$namespace"})'
# '(sum(up{job="apiserver"} == 1) / count(up{job="apiserver"})) * 100'
# '(sum(up{job="kube-controller-manager"} == 1) / count(up{job="kube-controller-manager"})) * 100'
# '(sum(up{job="kube-scheduler"} == 1) / count(up{job="kube-scheduler"})) * 100'
# '(1 - (sum(rate(apiserver_request_total{code=~"5.."}[5m])) or vector(0))/ sum(rate(apiserver_request_total[5m]))) * 100'
# '(sum(up{job="etcd"}==1)/count(up{job="etcd"}))*100'
# 'sum(sum by (cpu) (rate(node_cpu_seconds_total{job="node-exporter", mode!="idle"}[1m])))'
# 'sum(sum by (cpu) (rate(node_cpu_seconds_total{job="node-exporter"}[1m])))'
# 'sum(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)'
# 'sum(node_memory_MemTotal_bytes)'
# '(sum(node_filesystem_size_bytes) - sum(node_filesystem_free_bytes))'
# 'sum(node_filesystem_size_bytes)'
# 'sum(rate(container_network_receive_bytes_total{container="POD",pod!=""}[5m]))'
# 'sum(rate(container_network_transmit_bytes_total{container="POD",pod!=""}[5m]))'
# 'count(kube_pod_info)'
# 'sum(kube_pod_info{host_ip="$node_ip"})'
# 'node_memory_MemTotal_bytes{instance="$node_ip"}'
# 'node_memory_MemTotal_bytes{instance="$node_ip"} - node_memory_MemAvailable_bytes{instance="$node_ip"}'
# 'instance:node_cpu:rate:sum{instance="$node_ip"}'
# 'count(node_cpu_seconds_total{job="node-exporter",mode="idle",instance="$node_ip"}) by(instance)'
# 'instance:node_network_transmit_bytes:rate:sum{instance="$node_ip"}'
# 'instance:node_network_receive_bytes:rate:sum{instance="$node_ip"}'
# 'node_filesystem_size_bytes{mountpoint="/",instance="$node_ip"}'
# 'node_filesystem_size_bytes{mountpoint="/",instance="$node_ip"} - node_filesystem_free_bytes{mountpoint="/",instance="$node_ip"}'
# 'sum(container_memory_working_set_bytes{pod="$pod_name",namespace="$namespace",container="",}) BY (pod, namespace)'
# 'pod:container_cpu_usage:sum{pod="$pod_name",namespace="$namespace"}'
# 'pod:container_fs_usage_bytes:sum{pod="$pod_name",namespace="$namespace"}'
# 'sum(irate(container_network_receive_bytes_total{pod="$pod_name", namespace="$namespace"}[5m])) by (pod, namespace)'
# 'sum(irate(container_network_transmit_bytes_total{pod="$pod_name", namespace="$namespace"}[5m])) by (pod, namespace)'