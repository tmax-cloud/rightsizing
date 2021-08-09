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