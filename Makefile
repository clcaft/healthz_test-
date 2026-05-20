SELF_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
SELF_DIR := $(dir $(SELF_PATH))

include .env.example
export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

GOLANGCI_LINT_VERSION:=v1.61.0

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker-compose
	docker-compose up --build -d postgres rabbitmq && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

swag-v1: ### swag init
	swag init -g internal/controller/http/v1/router.go
.PHONY: swag-v1

run: swag-v1 ### swag run
	go mod tidy && go mod download && \
	DISABLE_SWAGGER_HTTP_HANDLER='' GIN_MODE=debug CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

lint: ### check by golangci linter
	$(LOCAL_BIN)/golangci-lint run --fix
.PHONY: linter-golangci

lint-docker:
	docker run --rm -v $(SELF_DIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v --timeout 10m
.PHONY: linter-golanci-docker

linter-hadolint: ### check by hadolint linter
	#git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
	docker run --rm -i hadolint/hadolint < Dockerfile
.PHONY: linter-hadolint

linter-dotenv: ### check by dotenv linter
	dotenv-linter
.PHONY: linter-dotenv

test: ### run test
	go test -v -cover -race ./internal/...
.PHONY: test

mock: ### run mockgen
	mockgen -source ./internal/usecase/interfaces.go -package mocks > ./internal/usecase/mocks/mocks.go
	mockgen -source ./internal/repo/interfaces.go -package mocks > ./internal/repo/mocks/mocks.go
.PHONY: mock

bin-deps-linters:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_LINT_VERSION)
	curl -sSfL https://raw.githubusercontent.com/dotenv-linter/dotenv-linter/master/install.sh | sh -s
.PHONY: bin-deps-linters

bin-deps:
	GOBIN=$(LOCAL_BIN) go install mvdan.cc/gofumpt@latest
	GOBIN=$(LOCAL_BIN) go install github.com/daixiang0/gci@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
.PHONY: bin-deps

cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out
.PHONY: cover

format:
	# ~/go/bin/gofumpt -extra .
	find . -name '*.go' -exec $(LOCAL_BIN)/gofumpt -extra -w {} +
	$(LOCAL_BIN)/gci write . --skip-generated -s standard -s default -s "prefix(gl.eda1.ru)"
.PHONY: format

proto-gen:
	# sudo apt install -y protobuf-compiler
	protoc -I internal/proto example.proto --go_out=internal/proto/go --go_opt=paths=source_relative --go-grpc_out=internal/proto/go --go-grpc_opt=paths=source_relative
.PHONY: proto-gen
