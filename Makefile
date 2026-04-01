-include .env
GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
SQLC := sqlc
PROJECT_NAME := auth-storage
ENV := dev
DOCKER := docker
DOCKER_COMPOSE_BIN := ENV=$(ENV) PROJECT_NAME=$(PROJECT_NAME) docker compose
IMAGE_NAME := anhle3532/auth-storage
IMAGE_TAG  := dev-latest

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:./docs \
	       $(API_PROTO_FILES)

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: wire
# regenerate wire
wire:
	go run -mod=mod github.com/google/wire/cmd/wire ./cmd/server/

.PHONY: test-e2e
# run e2e tests
test-e2e:
	go test -v ./test/e2e/...

.PHONY: all
# generate all
all:
	make api
	make config
	make generate

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# sqlc
.PHONY: sqlc
sqlc:
	$(SQLC) generate

.PHONY: sqlc-check
sqlc-check:
	$(SQLC) vet

.PHONY: sqlc-install
sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: migrate
migrate: timescale
	@$(DOCKER_COMPOSE_BIN) up migrate

redo-migrate: timescale
	@$(DOCKER_COMPOSE_BIN) run --rm migrate -path /migrations/sql drop
	@$(DOCKER_COMPOSE_BIN) up migrate

.PHONY: timescale
timescale:
	@$(DOCKER_COMPOSE_BIN) up timescale -d

# Docker login using environment variables
.PHONY: docker-login
docker-login:
	@if [ -z "$(DOCKER_USERNAME)" ]; then \
		echo "Error: DOCKER_USERNAME is not set. Please set it in your environment or a .env file."; \
		exit 1; \
	fi
	@if [ -z "$(DOCKER_PASSWORD)" ]; then \
		echo "Error: DOCKER_PASSWORD is not set. Please set it in your environment or a .env file."; \
		exit 1; \
	fi
	@echo "$(DOCKER_PASSWORD)" | docker login \
		--username $(DOCKER_USERNAME) \
		--password-stdin

.PHONY: docker-build
docker-build: 
	@$(DOCKER) build --rm --force-rm -t $(IMAGE_NAME):$(IMAGE_TAG) .
	-@$(DOCKER) rmi ${docker images -f "dangling=true" -q}

docker-push: docker-build docker-login
	@$(DOCKER) push $(IMAGE_NAME):$(IMAGE_TAG)

key-pair:
	@mkdir -p ./cert
	@KID="$$(openssl rand -hex 16)"; \
	echo "Generating RSA key pair for $$KID..."; \
	openssl genpkey -algorithm RSA -out ./cert/$$KID-private.pem -pkeyopt rsa_keygen_bits:2048 2>/dev/null; \
	openssl rsa -pubout -in ./cert/$$KID-private.pem -out ./cert/$$KID-public.pem 2>/dev/null; \
	echo $$KID > ./cert/active_kid.txt; \
	echo "✅ Keys generated! Check the ./cert directory."; \
	echo "🔑 Active pointer updated to: $$KID"

clear-key-pair:
	@rm ./cert/*

.PHONY: swagger
swagger:
	@$(DOCKER_COMPOSE_BIN) up swagger-ui -d
	@echo "Swagger UI is running at http://localhost:8080"

.PHONY: vault-up
vault-up:
	@echo "Generating fresh Root Token..."
	@$(eval DEV_TOKEN := $(shell openssl rand -hex 16))
	@echo "VAULT_DEV_ROOT_TOKEN_ID=$(DEV_TOKEN)" > .vault.env
	@echo "VAULT_ADDR=http://0.0.0.0:8200" >> .vault.env
	@echo "--------------------------------------------------"
	@echo "🚀 Starting Vault with Root Token: $(DEV_TOKEN)"
	@echo "--------------------------------------------------"
	@docker-compose --env-file .vault.env up -d

.PHONY: init-vault
init-vault:
	@echo "Checking Vault KV-V2 Secrets Engine..."
	@# Load the token from .env
	@export $$(grep -v '^#' .env | xargs) && \
	if docker exec -e VAULT_TOKEN=$${VAULT_DEV_ROOT_TOKEN_ID} -e VAULT_ADDR=http://127.0.0.1:8200 auth-storage-hashicorp vault secrets list | grep -q "^secret/"; then \
		echo "✅ Vault is already initialized at /secret. Skipping."; \
	else \
		echo "Initializing Vault KV-V2..."; \
		docker exec -e VAULT_TOKEN=$${VAULT_DEV_ROOT_TOKEN_ID} -e VAULT_ADDR=http://127.0.0.1:8200 \
		auth-storage-hashicorp vault secrets enable -path=secret kv-v2 && \
		echo "✅ Vault successfully initialized."; \
	fi

.PHONY: create-m2m-vault
create-m2m-vault:
	@if [ -z "$(id)" ] || [ -z "$(name)" ]; then \
		echo "❌ Error: Missing required arguments."; \
		echo "Usage: make create-m2m-vault id=\"service-id\" name=\"Service Name\""; \
		exit 1; \
	fi
	@echo "Generating credentials and pushing to Vault..."
	@# Load the token from the standard .env file we created
	@export $$(grep -v '^#' .env | xargs) && \
	 go run scripts/create_m2m_client.go -id "$(id)" -name "$(name)" -vault