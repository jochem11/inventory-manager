# inventory-manager: run `make` to list the targets. Each service has its own
# Makefile too (make -C services/<name>).

# Settings and credentials (DB_PASSWORD, …) from .env, so the MySQL tests run.
-include .env
export

# The Go modules, all part of go.work.
MODULES  := shared services/user-service services/auth-service services/item-service services/graphql-gateway
SERVICES := services/user-service services/auth-service services/item-service services/graphql-gateway
PACKAGES := $(foreach m,$(MODULES),./$(m)/...)

.DEFAULT_GOAL := help
.PHONY: help test vet proto typecheck check

help: ## List the targets
	@grep -hE '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

test: ## Run the Go tests of every module (MySQL tests use .env and skip without MySQL)
	go test $(PACKAGES)

vet: ## Vet every Go module
	go vet $(PACKAGES)

proto: ## Regenerate the gRPC code of every service
	@for s in $(SERVICES); do $(MAKE) -s -C $$s proto || exit 1; done

typecheck: ## Type-check the web app
	cd web && bun run typecheck

check: vet test typecheck ## Everything a change should pass: vet, tests, type-check
