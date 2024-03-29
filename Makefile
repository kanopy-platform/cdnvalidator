GO_MODULE := $(shell git config --get remote.origin.url | grep -o 'github\.com[:/][^.]*' | tr ':' '/')
CMD_NAME := $(shell basename ${GO_MODULE})
DEFAULT_APP_PORT ?= 8080
DEFAULT_CONFIG_FILE := /config/example_config.yaml
ENV_VARIABLES := -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_SESSION_TOKEN # ENV to pass into container
LOCAL_REGISTRY := registry.example.com
SWAGGER := docker run --rm -it -e GOPATH=${GOPATH}:/go -v ${HOME}:${HOME} -w ${PWD} quay.io/goswagger/swagger

RUN ?= .*
PKG ?= ./...
.PHONY: test
test: ## Run tests in local environment
	golangci-lint run --timeout=5m $(PKG)
	go test -short -cover -race -run=$(RUN) $(PKG)

.PHONY: test-integration
test-integration: ## Run integration tests
# Need to pass in args. Example:
# make test-integration DISTRIBUTION="<distribution>" PATHS="<comma separated paths>"
# make test-integration DISTRIBUTION="<distribution>" PATHS="<comma separated paths>" ACCESS_ID="<aws id>" ACCESS_SECRET="<aws secret>"
	go test -run=Integration -v ./.../cloudfront -args -distribution=$(DISTRIBUTION) -paths=$(PATHS) -access-id=$(ACCESS_ID) -access-secret=$(ACCESS_SECRET)

.PHONY: docker-build-test
docker-build-test: ## Build local development docker image with cached go modules, builds, and tests
	@docker build -f build/Dockerfile-test -t $(CMD_NAME)-test:latest .

.PHONY: docker-test
docker-test: docker-build-test ## Run tests using local development docker image
	@docker run -v $(shell pwd):/go/src/$(GO_MODULE):delegated $(CMD_NAME)-test make test RUN=$(RUN) PKG=$(PKG)

.PHONY: docker-snyk
docker-snyk: ## Run local snyk scan, SNYK_TOKEN environment variable must be set
	@docker run --rm -e SNYK_TOKEN -w /go/src/$(GO_MODULE) -v $(shell pwd):/go/src/$(GO_MODULE):delegated snyk/snyk:golang

.PHONY: docker
docker:
	@docker build -t $(CMD_NAME):latest .

.PHONY: docker-run
docker-run: docker ## Build and run the application in a local docker container
	@docker run -p ${DEFAULT_APP_PORT}:${DEFAULT_APP_PORT} -v ${PWD}${DEFAULT_CONFIG_FILE}:${DEFAULT_CONFIG_FILE} ${ENV_VARIABLES} $(CMD_NAME):latest --config-file=${DEFAULT_CONFIG_FILE}

.PHONY: docker-push-local
docker-push-local: ## Build and push the image to local docker registry
	@docker build -t $(LOCAL_REGISTRY)/$(CMD_NAME):latest .
	@docker push $(LOCAL_REGISTRY)/$(CMD_NAME):latest

.PHONY: swagger-gen
swagger-gen: ## Generate swagger OpenAPI specification
	$(SWAGGER) generate spec -o ./swagger/swagger.json --scan-models

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
