BIN := $(abspath ./bin)
BUILD := build
GO ?= go
GO_ENV ?= GOBIN=$(BIN) CGO_ENABLE=0

$(BIN)/stringer:
	$(GO_ENV) $(GO) install -mod=mod golang.org/x/tools/cmd/stringer@v0.1.4

$(BIN)/wire:
	$(GO_ENV) $(GO) get github.com/google/wire/cmd/wire@v0.5.0

$(BUILD)/server:
	$(GO_ENV) go build -o $(BUILD)/server ./cmd/server

.PHONY: clean
clean:
	$(RM) -r ./build ./bin

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
	@PATH=$(BIN):${PATH} $(GO_ENV) $(GO) generate ./...

.PHONY: scan
scan: $(BUILD)/server
	trivy fs -s "HIGH,CRITICAL" --ignore-unfixed --exit-code 1 $(BUILD)
