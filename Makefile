.PHONY: fmt vet lint test test-coverage build check clean install generate docs-serve

HOSTNAME=registry.terraform.io
NAMESPACE=cdot65
NAME=prisma-airs
BINARY=terraform-provider-${NAME}
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
VERSION=0.1.0

export GOPRIVATE=github.com/cdot65/*

## Format source code
fmt:
	gofmt -s -w .

## Run go vet
vet:
	go vet ./...

## Run golangci-lint
lint:
	golangci-lint run ./...

## Run tests with race detector
test:
	go test -race ./...

## Run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## Run acceptance tests (requires real credentials)
testacc:
	TF_ACC=1 go test -race ./... -v -timeout 120m

## Build the provider binary
build:
	go build -o ${BINARY}

## Run all checks (CI equivalent)
check: fmt vet lint test

## Install the provider locally for development
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

## Generate Terraform provider documentation
generate:
	go generate ./...

## Remove build artifacts
clean:
	rm -f ${BINARY}
	rm -f coverage.out
	rm -rf site/
	rm -rf dist/

## Serve MkDocs locally
docs-serve:
	mkdocs serve
