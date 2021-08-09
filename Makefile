run:
	uvicorn main:app --reload --host=0.0.0.0 --port=8000

docker-build: build ## docker build
	ifdef TAG
		docker build -t $(IMAGE_NAME):$(TAG)
	else
		docker build -t $(IMAGE_NAME)
	endif

build: poetry-update poetry-export

poetry-update: ## poetry update
	poetry update

poetry-export: ## export requirements.txt from poetry
	poetry export -f requirements.txt --output requirements.txt


.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'