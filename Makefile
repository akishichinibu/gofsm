GOLANGCI_VERSION := v2.1.5

init/golangci-lint:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_VERSION)

build:
	@go build

test:
	@go test ./...

lint:
	@$(shell go env GOPATH)/bin/golangci-lint run ./... --timeout 5m
