from datetime import datetime
import logging
from typing import Any, Optional
import re

import pandas as pd
from pydantic import BaseModel, Field, validator
from sklearn.preprocessing import StandardScaler

logger = logging.getLogger(__name__)

regex = re.compile('^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$')


class AnalysisQuery(BaseModel):
    name: str = Field(..., description="The name of kubernetes object (pod)")
    namespace: Optional[str] = Field(None, example="default", description="The namespace of kubernetes object")
    url: str = Field(None, example="http://localhost:9090", description="The url of prometheus")
    description: Optional[str] = Field(None, example="Kubernetes deployment analysis")

    @classmethod
    @validator("url")
    def validate_url(cls, v):
        if not v.startswith("https://") and not v.startswith("http://"):
            return False
        return True

    @classmethod
    @validator("name")
    def name_rfc1123(cls, v):
        return regex.match(v) is not None


class QueryParams:
    def __init__(self, query: Optional[AnalysisQuery] = None, data: Any = None):
        self._query = query
        self.request_time = datetime.now()

        self._data = data
        self._scaler = StandardScaler()
        self._scaler.fit(self._data)

        self._scaled_data = pd.DataFrame(self._scaler.transform(self._data),
                                         columns=self._data.columns,
                                         index=self._data.index)

        logger.info(f"Successfully create QueryParams({query})")

    @property
    def data(self):
        return self._data

    @property
    def scaled_data(self):
        return self._scaled_data

    @property
    def url(self):
        return self._query.url

    @property
    def namespace(self):
        return self._query.namespace

    @property
    def name(self):
        return self._query.name

    def inverse_transform(self, data: pd.Series, field: str) -> pd.Series:
        scaler = StandardScaler()
        scaler.fit(self._data[[field]])
        return pd.Series(scaler.inverse_transform(data), index=data.index)

    def transform_column(self, field: str):
        if field not in self.data:
            return None, None

        scaler = StandardScaler()
        scaler.fit(self._data[[field]])

        scaled_data = scaler.transform(self.data[[field]])
        return scaler, scaled_data
