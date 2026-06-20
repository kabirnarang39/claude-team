.PHONY: dev test build install clean

dev: ## Start Anton with live reload
	cd mcp && npm install --silent && cd ..
	go run main.go

test: ## Run all tests
	go test ./...

build: ## Build binary for current platform
	go build -o anton .

install: ## Install binary to /usr/local/bin
	go build -o /usr/local/bin/anton .

clean: ## Remove built binary and runtime state
	rm -f anton
	rm -rf .claude-team/

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
