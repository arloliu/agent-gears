MODULE  := github.com/arloliu/agent-gears
BINARY  := agent-gears
CMD     := ./cmd/agent-gears
BIN_DIR := bin

.PHONY: all
all: build

.PHONY: build
build: ## Build the agent-gears binary into bin/
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY) $(CMD)

.PHONY: install
install: ## Build and install agent-gears to $GOBIN (or $GOPATH/bin)
	go install $(CMD)

.PHONY: run
run: ## Run agent-gears from source (pass args via ARGS="...")
	go run $(CMD) $(ARGS)

.PHONY: test
test: ## Run the test suite
	go test ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format all Go source files
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any Go source file is not gofmt-formatted
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

.PHONY: check
check: fmt-check vet test ## Run fmt-check, vet, and test

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

.PHONY: help
help: ## Show this help message
	@grep -hE '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-12s %s\n", $$1, $$2}'
