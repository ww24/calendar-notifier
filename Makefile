BIN := $(abspath ./bin)
GO ?= go
GO_ENV ?= GOBIN=$(BIN)

$(BIN)/stringer:
	$(GO_ENV) $(GO) install -mod=mod golang.org/x/tools/cmd/stringer

$(BIN)/wire:
	$(GO_ENV) $(GO) install -mod=mod github.com/google/wire/cmd/wire

.PHONY: build
build:
	docker build -t calendar-notifier .

.PHONY: run
run: CONFIG_BASE64 ?= $(shell base64 < config.yml)
run: SERVICE_ACCOUNT ?= $(shell base64 < service_account.json)
run:
	@CONFIG_BASE64=$(CONFIG_BASE64) \
	docker-compose up -d
	docker logs -f calendar-notifier

.PHONY: generate
generate: $(BIN)/stringer $(BIN)/wire
	PATH=$(BIN):${PATH} $(GO_ENV) $(GO) generate ./...
