# Development tasks for Cahut. Run "make" to list them.

# Keep these in step with the versions pinned in .github/workflows.
GOLANGCI_LINT_VERSION := v2.15.0
GORELEASER_VERSION    := v2.18.3

# The tools run from the Go module cache, so nothing needs installing.
# Set GOLANGCI_LINT=golangci-lint or GORELEASER=goreleaser to use your own.
GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GORELEASER    ?= go run github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

.DEFAULT_GOAL := help

# The sites need Node.js at the version of .nvmrc: run "nvm use" first.
SITES_DEPS := sites/node_modules/.package-lock.json

.PHONY: help build dev test lint fmt check snapshot release-check sites site-dev docs-dev clean

help: ## List the tasks
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/cahut for this machine
	CGO_ENABLED=0 go build -trimpath -o bin/cahut ./cmd/cahut

dev: ## Serve the Cahuts of tmp/dev on http://localhost:8080
	go run ./cmd/cahut serve --dev --data tmp/dev

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

sites: $(SITES_DEPS) ## Build cahut.cloud and docs.cahut.cloud into sites/*/dist
	cd sites && npm run build

site-dev: $(SITES_DEPS) ## Serve cahut.cloud on http://localhost:4321
	cd sites && npm run dev --workspace cahut.cloud

docs-dev: $(SITES_DEPS) ## Serve docs.cahut.cloud on http://localhost:4322
	cd sites && npm run dev --workspace docs.cahut.cloud

$(SITES_DEPS): sites/package.json sites/package-lock.json $(wildcard sites/*/package.json)
	cd sites && npm ci
	touch $@

clean: ## Remove the build output and the development data
	rm -rf bin dist tmp sites/node_modules sites/*/dist sites/*/.astro
