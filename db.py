import os

# from aioinflux import InfluxDBClient
from influxdb import DataFrameClient
from fastapi import Request


class DBConnection:
    _client: DataFrameClient = None
    _bucket: str = ""

    def __init__(self):
        self.init()

    def init(self):
        if self._client is None:
            self._client = DataFrameClient(
                host=os.getenv('INFLUXDB_URL'),
                port=os.getenv('INFLUXDB_PORT'),
                username=os.getenv('INFLUXDB_USER'),
                password=os.getenv('INFLUXDB_SECRET'),
                database=os.getenv('INFLUXDB_BUCKET'))
                # output='dataframe',
                # mode='blocking')

    def close(self):
        self._client.close()

    @property
    def connection(self):
        return self._client

    @property
    def query(self) -> callable:
        return self._client.query


def get_db(request: Request):
    return request.app.state.db