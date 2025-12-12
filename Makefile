.DEFAULT_GOAL:=help
-include .makerc

# --- Config -----------------------------------------------------------------

GOMODS=$(shell find . -type f -name go.mod)
# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .lefthook
	@:

.PHONY: .mise
# Install dependencies
.mise: msg := $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)$$ brew update$(br)$$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html$(br)$(br)
.mise:
ifeq (, $(shell command -v mise))
	$(error ${msg})
endif
	@mise install

# Configure git hooks for lefthook
.lefthook:
	@lefthook install

### Tasks

.PHONY: check
## Run lint & tests
check: tidy lint test

.PHONY: tidy
## Run go mod tidy
tidy:
	@echo "〉go mod tidy"
	@go mod tidy

.PHONY: lint
## Run linter
lint:
	@echo "〉golangci-lint run"
	@biome check
	@golangci-lint run

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@echo "〉golangci-lint run fix"
	@biome check --write
	@golangci-lint run --fix

.PHONY: test
## Run tests
test:
	@echo "〉go test"
	@# see https://github.com/pterm/pterm/issues/482
	@GO_TEST_TAGS=-skip go test -tags=safe -coverprofile=coverage.out
	@#GO_TEST_TAGS=-skip go test -tags=safe -coverprofile=coverage.out -race

.PHONY: test.update
## Run tests
test.update:
	@echo "〉go test"
	@GO_TEST_TAGS=-skip go test -tags=safe -coverprofile=coverage.out -update

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@echo "〉go mod outdated"
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: install
## Install binary
install:
	@echo "〉installing ${GOPATH}/bin/squadron"
	@go build -tags=safe -o ${GOPATH}/bin/squadron cmd/main.go

.PHONY: build
## Build binary
build:
	@mkdir -p bin
	@echo "〉building bin/squadron"
	@go build -tags=safe -o bin/squadron cmd/main.go

### Documentation

.PHONY: docs
## Open docs
docs: docs.cli
	@echo "〉starting docs"
	@cd docs && bun install
	@cd docs && bun run dev

.PHONY: docs.cli
## Generate cli
docs.cli:
	@echo "〉generating cli reference"
	@go run scripts/cmd-reference/main.go

.PHONY: godocs
## Open go docs
godocs:
	@echo "〉starting go docs"
	@go doc -http

### Utils

.PHONY: help
## Show help text
help:
	@echo "Squadron | Compose for Kubernetes\n"
	@echo "Usage:\n  make [task]"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "%-23s %s\n\n", cmd, help; help=""; \
			printf "\n%s:\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  %-23s %s\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                        " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        " help "\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)
