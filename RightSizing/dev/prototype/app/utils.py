from fastapi_cache import FastAPICache


def redis_key_builder(
    func,
    url: str = None,
    namespace: str = None,
    name: str = None,
    *args,
    **kwargs,
):
    prefix = FastAPICache.get_prefix()
    cache_key = f"{prefix}:{url}:{namespace}:{name}:{func.__module__}:{func.__name__}:{args}:{kwargs}"
    return cache_key
