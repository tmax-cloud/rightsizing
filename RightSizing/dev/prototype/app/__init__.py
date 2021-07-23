from celery import Celery
from fastapi import FastAPI
from fastapi.middleware.wsgi import WSGIMiddleware

import constants
from app import router
from app.plotly_dash import create_dash_app


def make_celery(app):
    celery = Celery(
        app.import_name,
        broker=app.config[constants.CELERY_BROKER_URL],
        backend=app.config[constants.CELERY_BACKEND_URL]
    )
    celery.conf.update(app.config)

    class ContextTask(celery.Task):
        def __call__(self, *args, **kwargs):
            with app.app_context():
                return self.run(*args, **kwargs)
    celery.Task = ContextTask
    return celery


def create_app():
    app = FastAPI()
    # app.config.update(config)
    dash_app = create_dash_app('/dash/')

    app.include_router(router.router)
    app.mount('/', WSGIMiddleware(dash_app.server))
    return app


