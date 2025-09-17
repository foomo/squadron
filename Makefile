.DEFAULT_GOAL:=help
-include .makerc

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise
	@:

.PHONY: .mise
# Install dependencies
.mise:
	@mise install

### Tasks

.PHONY: check
## Run tests and linters
check: tidy lint test

.PHONY: doc
## Run tests
doc:
	@open "http://localhost:6060/pkg/github.com/foomo/squadron/"
	@godoc -http=localhost:6060 -play

.PHONY: test
## Run tests
test:
	@# see https://github.com/pterm/pterm/issues/482
	@GO_TEST_TAGS=-skip go test -tags=safe -coverprofile=coverage.out
	@#GO_TEST_TAGS=-skip go test -tags=safe -coverprofile=coverage.out -race

.PHONY: lint
## Run linter
lint:
	@golangci-lint run

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@golangci-lint run --fix

.PHONY: tidy
## Run go mod tidy
tidy:
	@go mod tidy

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: install
## Install binary
install:
	@echo "installing ${GOPATH}/bin/squadron"
	@go build -tags=safe -o ${GOPATH}/bin/squadron cmd/main.go

.PHONY: build
## Build binary
build:
	@mkdir -p bin
	@echo "building bin/squadron"
	@go build -tags=safe -o bin/squadron cmd/main.go

### Utils

.PHONY: help
## Show help text
help:
	@echo "\033[1;36mSquadron | Compose for Kubernetes\033[0m"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "\033[36m%-23s\033[0m %s\n\n", cmd, help; help=""; \
			printf "\n\033[1;36m%s\033[0m\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  \033[36m%-23s\033[0m %s\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                        " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        " help "\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)

