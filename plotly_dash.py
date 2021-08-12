import asyncio
import os

import dash
import dash_core_components as dcc
import dash_html_components as html
from dash.dependencies import Input, Output, State
import flask
import requests
from fbprophet.plot import plot_plotly, plot_components_plotly

import constants
from query import query_and_analyze_pod
from models import AnalysisQuery

import ML

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
            html.Div([dcc.Graph(id='cpu_forecast')]),
            html.Div([dcc.Graph(id='memory_forecast')])
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
        Output('memory_forecast', 'figure'),
        Input('prometheus-url', 'data'),
        Input('pods', 'value')
    )
    def update_graph(url: str, name: str):
        if not name:
            return [dash.no_update for i in range(2)]
        os.environ['HOST'] = url
        namespace, name = name.split()
        data = asyncio.run(query_and_analyze_pod(AnalysisQuery(namespace=namespace, name=name)))

        scaled_data = data.data

        cpu_df = scaled_data[constants.CPU_FIELD].to_frame().reset_index().rename(
            columns= {"index": "ds", constants.CPU_FIELD: "y"})
        cpu_m, cpu_forecast = ML.forecasting(cpu_df)
        memory_df = scaled_data[constants.MEMORY_FIELD].to_frame().reset_index().rename(
            columns= {"index": "ds", constants.MEMORY_FIELD: "y"})
        memory_m, memory_forecast = ML.forecasting(memory_df)
        return [plot_plotly(cpu_m, cpu_forecast), plot_plotly(memory_m, memory_forecast)]

    return app