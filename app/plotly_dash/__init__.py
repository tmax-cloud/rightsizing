import asyncio
import dash
import dash_core_components as dcc
import dash_html_components as html
from dash.dependencies import Input, Output, State
from fastapi.logger import logger
import flask
import pandas as pd
import plotly.graph_objs as go
from sklearn.preprocessing import StandardScaler
import requests

import constants
from app.router import query_and_analyze_pod
from app.models import AnalysisQuery

import ML

pd.options.plotting.backend = "plotly"

external_stylesheets = [
    # Dash CSS
    'https://codepen.io/chriddyp/pen/bWLwgP.css',
    # Loading screen CSS
    'https://codepen.io/chriddyp/pen/brPBPO.css'
]


def create_dash_app(routes_pathname_prefix: str):
    server = flask.Flask(__name__)

    app = dash.Dash(__name__,
                    server=server,
                    routes_pathname_prefix=routes_pathname_prefix,
                    external_stylesheets=external_stylesheets)

    app.layout = html.Div([
        html.H1(children="Rightsizing"),
        html.Div(["Prometheus URL: ",
                  dcc.Input(id='url', value='', type='text', autoFocus=True),
                  html.Button(id='submit', n_clicks=0, children='Submit')]),
        html.Hr(),
        html.Div([
            dcc.Dropdown(
                id='pods',
                options=[],
                value=None
            )
        ]),
        html.Div([
            html.Div([dcc.Graph(id='cpu_forecast'),
                      dcc.Graph(id='cpu_abnormal')]),
            html.Div([dcc.Graph(id='memory_forecast'),
                      dcc.Graph(id='memory_abnormal')])
        ]),
        dcc.Store(id='prometheus-url', data='string')
    ])

    @app.callback(
        Output('pods', 'options'),
        Output('prometheus-url', 'data'),
        Input('submit', 'n_clicks'),
        State('url', 'value'),
    )
    def update_prometheus_url(n_clicks, url: str):
        if n_clicks == 0:
            return dash.no_update, dash.no_update

        if url.endswith('/'):
            url = url[:-1]
        url = url + constants.QUERY_ENDPOINT

        query = {
            'query': constants.FIELDS[0],
        }

        r = requests.get(url, params=query)
        result = r.json()['data']['result']
        pods = set()
        for pod in result:
            namespace = pod['metric'].get('namespace', None)
            name = pod['metric'].get('pod', None)
            if namespace and name:
                pods.add((namespace, name))
        pods = [{'label': f'namespace: {namespace}, pod: {name}', 'value': f'{namespace} {name}'}
                for namespace, name in pods]
        return pods, url

    @app.callback(
        Output('cpu_forecast', 'figure'),
        Output('cpu_abnormal', 'figure'),
        Output('memory_forecast', 'figure'),
        Output('memory_abnormal', 'figure'),
        Input('prometheus-url', 'data'),
        Input('pods', 'value')
    )
    def update_graph(url: str, name: str):
        logger.info(f'name: {name}')
        if not name:
            return [dash.no_update for i in range(4)]
        namespace, name = name.split()
        data = asyncio.run(query_and_analyze_pod(AnalysisQuery(url=url, namespace=namespace, name=name)))
        # if not data.any():
        #     return [dash.no_update for i in range(4)]

        scaler = StandardScaler()
        scaler.fit(data)

        cpu_scaler = StandardScaler()
        cpu_scaler.fit(data[[constants.CPU_FIELD]])
        memory_scaler = StandardScaler()
        memory_scaler.fit(data[[constants.MEMORY_FIELD]])

        scaled_df = pd.DataFrame(scaler.transform(data), columns=data.columns, index=data.index)

        cpu_forecast = ML.forecasting(scaled_df, constants.CPU_FIELD)
        cpu_forecast = pd.Series(cpu_scaler.inverse_transform(cpu_forecast), index=cpu_forecast.index)
        cpu_forecast = data[constants.CPU_FIELD].append(cpu_forecast)
        data['cpu_abnormal'] = ML.abnormal_detection(scaled_df, constants.CPU_FIELD)
        cpu_abnormal = data.loc[data['cpu_abnormal'] == -1, [constants.CPU_FIELD]]

        memory_forecast = ML.forecasting(scaled_df, constants.MEMORY_FIELD)
        memory_forecast = pd.Series(memory_scaler.inverse_transform(memory_forecast), index=memory_forecast.index)
        memory_forecast = data[constants.MEMORY_FIELD].append(memory_forecast)
        data['memory_abnormal'] = ML.abnormal_detection(scaled_df, constants.MEMORY_FIELD)
        memory_abnormal = data.loc[data['memory_abnormal'] == -1, [constants.MEMORY_FIELD]]

        cpu_forecast_fig = cpu_forecast.plot(title="CPU Usage Forecast")
        cpu_forecast_fig.add_vline(x=data.index[-1], line_dash="dot", line_color="red")

        cpu_abnormal_fig = data[constants.CPU_FIELD].plot(title="CPU Usage abnormal detection")
        cpu_abnormal_fig.add_trace(
            go.Scatter(x=cpu_abnormal.index, y=cpu_abnormal[constants.CPU_FIELD], mode="markers"))

        memory_forecast_fig = memory_forecast.plot(title="Memory Usage Forecast")
        memory_forecast_fig.add_vline(x=data.index[-1], line_dash="dot", line_color="red")

        memory_abnormal_fig = data[constants.MEMORY_FIELD].plot(title="Memory Usage abnormal detection")
        memory_abnormal_fig.add_trace(
            go.Scatter(x=memory_abnormal.index, y=memory_abnormal[constants.MEMORY_FIELD], mode="markers"))

        return [cpu_forecast_fig, cpu_abnormal_fig, memory_forecast_fig, memory_abnormal_fig]

    return app
