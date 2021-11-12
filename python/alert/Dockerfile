FROM python:3.9-slim

ENV PYTHONFAULTHANDLER=1 \
    PYTHONHASHSEED=random \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN python -m pip install --upgrade pip

RUN python -m venv /venv
RUN . /venv/bin/activate

COPY requirements.txt /app

RUN python -m pip install -r requirements.txt

COPY . .

CMD [""]