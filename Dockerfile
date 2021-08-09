FROM python:3.9-slim

ENV PYTHONFAULTHANDLER=1 \
    PYTHONHASHSEED=random \
    PYTHONUNBUFFERED=1

WORKDIR /app

RUN python -m pip install --upgrade pip

COPY requirements.txt requirements.txt
RUN --mount=type=cache,target=/root/.cache \
  pip install -r requirements.txt

COPY . .

CMD [""]