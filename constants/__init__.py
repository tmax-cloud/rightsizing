QUERY_ENDPOINT = '/api/v1/query' # Prometheus query endpoint
QUERY_TERM = "7d" # Prometheus query term
# 필터링할 컨테이너 리스트
PAUSE_CONTAINER = "POD"  # Pause container name in Prometheus.

# Query fields to Prometheus
FIELDS = [
    'pod:container_cpu_usage:sum',
    'pod:container_memory_usage_bytes:sum',
    'pod:container_fs_usage_bytes:sum'
]


CPU_FIELD = 'pod:container_cpu_usage:sum'
MEMORY_FIELD = 'pod:container_memory_usage_bytes:sum'
# Forecasting frequency
FREQ = "30S"
FORECAST_STEP = 1440