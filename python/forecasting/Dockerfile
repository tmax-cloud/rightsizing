FROM python:3.9

ENV PYTHONFAULTHANDLER=1 \
    PYTHONHASHSEED=random \
    PYTHONUNBUFFERED=1

WORKDIR /app

# RUN apk add --no-cache jpeg-dev zlib-dev libstdc++
# RUN apk --update add --no-cache --virtual .build-deps \
#     build-base linux-headers gcc g++ gfortran musl-dev libffi-dev openssl-dev curl zeromq-dev rust cargo

RUN python -m pip install --upgrade pip

RUN python -m venv /venv
RUN . /venv/bin/activate

RUN python -m pip install poetry

COPY pyproject.toml /app/
COPY poetry.lock /app/

RUN poetry config virtualenvs.create false \
  && poetry install --no-root --no-dev  --no-interaction --no-ansi

# RUN apk --no-cache del --purge .build-deps

COPY . .

CMD [""]
