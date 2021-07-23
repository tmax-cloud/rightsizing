import os


class LocalCeleryConfig:
    CELERY_BROKER_URL = os.getenv('BROKER_URL', None)
    CELERY_RESULT_BACKEND = os.getenv('CELERY_RESULT_BACKEND', None)


class ProductionCeleryConfig:
    CELERY_BROKER_URL = os.getenv('BROKER_URL', None)
    CELERY_RESULT_BACKEND = os.getenv('CELERY_RESULT_BACKEND', None)
