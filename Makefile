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

.PHONY: help build dev test lint fmt check snapshot release-check console console-dev console-types sites site-dev docs-dev docs-reference clean

help: ## List the tasks
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# The console is built by Vite into internal/console/ui/dist, then embedded
# in the binary with the console build tag. Without the tag, as with a plain
# go build, the binary serves a notice in its place. It needs Node.js at the
# version of .nvmrc: run "nvm use" first.
CONSOLE_DIR  := internal/console/ui
CONSOLE_DEPS := $(CONSOLE_DIR)/node_modules/.package-lock.json

build: console ## Build bin/cahut for this machine, with the console
	CGO_ENABLED=0 go build -tags console -trimpath -o bin/cahut ./cmd/cahut

dev: console ## Serve the Cahuts of tmp/dev on http://localhost:8080, their console on 8081
	go run -tags console ./cmd/cahut serve --dev --data tmp/dev

console: $(CONSOLE_DEPS) ## Build the console into internal/console/ui/dist
	cd $(CONSOLE_DIR) && npm run build

console-dev: $(CONSOLE_DEPS) ## Serve the console with live reload on http://localhost:5173, next to make dev
	cd $(CONSOLE_DIR) && npm run dev

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

# The reference of the docs comes from the code. It needs Go only, and CI
# checks that it is committed.
DOCS_REFERENCE := sites/docs.cahut.cloud/src/content/docs/reference

docs-reference: ## Write the reference of the docs from the code
	go run ./internal/reference $(DOCS_REFERENCE)

# The TypeScript types of the console's API come from the code too. They
# need Go only, and CI checks that they are committed.
CONSOLE_TYPES := internal/console/ui/src/api.gen.ts

console-types: ## Write the TypeScript types of the console's API from the code
	go run ./internal/apitypes $(CONSOLE_TYPES)

$(CONSOLE_DEPS): $(CONSOLE_DIR)/package.json $(CONSOLE_DIR)/package-lock.json
	cd $(CONSOLE_DIR) && npm ci
	touch $@

$(SITES_DEPS): sites/package.json sites/package-lock.json $(wildcard sites/*/package.json)
	cd sites && npm ci
	touch $@

clean: ## Remove the build output and the development data
	rm -rf bin dist tmp sites/node_modules sites/*/dist sites/*/.astro $(CONSOLE_DIR)/node_modules $(CONSOLE_DIR)/dist
