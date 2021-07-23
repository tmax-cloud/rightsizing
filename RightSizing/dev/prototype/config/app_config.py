import os


class LocalAppConfig:
    ENV = "development"
    DEBUG = True
    SECRET_KEY = os.getenv('SECRET_KEY')


class ProductionAppConfig:
    ENV = "production"
    DEBUG = False
    SECRET_KEY = os.getenv('SECRET_KEY')