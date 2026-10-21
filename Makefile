# Development tasks for Cahut. Run "make" to list them.

# Keep these in step with the versions pinned in .github/workflows.
GOLANGCI_LINT_VERSION := v2.15.0
GORELEASER_VERSION    := v2.18.3

# The tools run from the Go module cache, so nothing needs installing.
# Set GOLANGCI_LINT=golangci-lint or GORELEASER=goreleaser to use your own.
GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GORELEASER    ?= go run github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

.DEFAULT_GOAL := help

.PHONY: help build test lint fmt check snapshot release-check clean

help: ## List the tasks
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/cahut for this machine
	CGO_ENABLED=0 go build -trimpath -o bin/cahut ./cmd/cahut

test: ## Run the tests with the race detector
	go test -race ./...

lint: ## Check go.mod and run the linters
	go mod tidy -diff
	$(GOLANGCI_LINT) run

fmt: ## Format the code
	$(GOLANGCI_LINT) fmt

check: lint test ## Run everything CI runs, before pushing

snapshot: ## Build the release archives into dist/, without publishing
	$(GORELEASER) release --snapshot --clean

release-check: ## Validate the GoReleaser configuration
	$(GORELEASER) check

clean: ## Remove the build output
	rm -rf bin dist
