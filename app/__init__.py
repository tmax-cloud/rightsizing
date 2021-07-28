from fastapi import FastAPI
from fastapi.middleware.wsgi import WSGIMiddleware

from .router import router
from .plotly_dash import create_dash_app


def create_app():
    app = FastAPI()

    app.include_router(router)

    dash_app = create_dash_app('/dash/')
    app.mount('/', WSGIMiddleware(dash_app.server))

    return app
