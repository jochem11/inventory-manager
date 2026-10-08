# user-service: run `make` to list the targets.

# Settings and credentials (DB_PASSWORD, …) from the repo's .env, for run and
# test; see .env.example. Missing .env: the services' defaults apply.
-include ../../.env
export

MODULE := github.com/jochem11/inventory-manager/services/user-service
PROTO  := ../../proto

# Each .proto gets its Go import path here (path;package), so the shared
# .proto files need no Go-specific go_package option.
USER_PB        := user/user.proto=$(MODULE)/pkg/pb/user;userpb
USER_EVENTS_PB := user/events.proto=$(MODULE)/pkg/pb/user;userpb
# Consumed: auth-service events (IdentityRegistered).
AUTH_EVENTS_PB := auth/events.proto=$(MODULE)/pkg/pb/auth;authpb
PROTOS         := user/user.proto user/events.proto auth/events.proto
GO_OPTS        := '--go_opt=M$(USER_PB)' '--go_opt=M$(USER_EVENTS_PB)' '--go_opt=M$(AUTH_EVENTS_PB)'
GRPC_OPTS      := '--go-grpc_opt=M$(USER_PB)' '--go-grpc_opt=M$(USER_EVENTS_PB)' '--go-grpc_opt=M$(AUTH_EVENTS_PB)'

# `make tools` installs the protoc plugins here, which may not be on PATH.
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.DEFAULT_GOAL := help
.PHONY: help tools proto run test

help: ## List the targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-8s %s\n", $$1, $$2}'

tools: ## Install the protoc plugins for Go
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: ## Regenerate pkg/pb from the .proto files
	rm -rf pkg/pb && mkdir -p pkg/pb
	protoc -I $(PROTO) \
		--go_out=pkg/pb      --go_opt=paths=source_relative      $(GO_OPTS) \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative $(GRPC_OPTS) \
		$(PROTOS)

run: ## Start the gRPC server on :50051
	go run ./cmd

test: ## Run the tests (repository tests need MySQL, see DB_* env, and skip without it)
	go test ./...
