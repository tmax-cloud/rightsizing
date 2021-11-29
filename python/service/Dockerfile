FROM python:3.9

ENV PYTHONFAULTHANDLER=1 \
    PYTHONHASHSEED=random \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN python -m venv /venv
RUN . /venv/bin/activate

RUN python -m pip install --upgrade pip
RUN python -m pip install poetry

COPY pyproject.toml /app/
COPY poetry.lock /app/

RUN poetry config virtualenvs.create false \
  && poetry install --no-root --no-dev  --no-interaction --no-ansi

COPY . .

CMD [""]
