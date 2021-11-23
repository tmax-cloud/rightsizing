import numpy as np


MARGIN = 0.2


def quantiles(data: np.ndarray, quantile: int) -> np.ndarray:
    m = np.percentile(data, quantile)
    return m


def extract_resource_usage(data: np.ndarray, quantile: int):
    percentiles = quantiles(data, quantile)
    return percentiles * (1 + MARGIN)