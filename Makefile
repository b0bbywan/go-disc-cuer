BINARY  := disc-cuer
MODULE  := github.com/b0bbywan/go-disc-cuer
GO      ?= go
DIST    := dist

# Version stamped into config.AppVersion at build time. Defaults to the git
# description (tag/sha, with -dirty when the tree has uncommitted changes), and
# falls back to "dev" outside a git checkout.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X $(MODULE)/config.AppVersion=$(VERSION)

# golangci-lint version, kept in sync with the CI workflow.
GOLANGCI_LINT_VERSION := v2.12.2

.PHONY: all deps build dist test lint vet tidy clean

all: build

## deps: install the libdiscid build dependency (Debian/Ubuntu or Fedora).
deps:
	@if command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y libdiscid-dev; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y libdiscid libdiscid-devel; \
	else \
		echo "unsupported package manager: install the libdiscid dev headers manually" >&2; \
		exit 1; \
	fi

## build: compile the binary with the version stamped in.
build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## dist: build the linux/amd64 release binary into $(DIST)/, named by OS/arch.
dist:
	mkdir -p $(DIST)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		$(GO) build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-amd64 .

## test: run the full suite with the race detector and coverage.
test:
	$(GO) test -race -coverprofile=coverage.out ./...

## lint: run golangci-lint (installs the pinned version if absent).
lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	golangci-lint run

## vet: run go vet across all packages.
vet:
	$(GO) vet ./...

## tidy: prune and sync go.mod / go.sum.
tidy:
	$(GO) mod tidy

## clean: remove build and coverage artifacts.
clean:
	rm -rf $(BINARY) coverage.out $(DIST)
