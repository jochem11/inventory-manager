# graphql-gateway: run `make` to list the targets.

# Settings and credentials (DB_PASSWORD, …) from the repo's .env, for run and
# test; see .env.example. Missing .env: the services' defaults apply.
-include ../../.env
export

MODULE := github.com/jochem11/inventory-manager/services/graphql-gateway
PROTO  := ../../proto

# Each .proto gets its Go import path here (path;package), so the shared
# .proto files need no Go-specific go_package option.
USER_PB   := user/user.proto=$(MODULE)/pkg/pb/user;userpb
AUTH_PB   := auth/auth.proto=$(MODULE)/pkg/pb/auth;authpb
PROTOS    := user/user.proto auth/auth.proto
GO_OPTS   := '--go_opt=M$(USER_PB)' '--go_opt=M$(AUTH_PB)'
GRPC_OPTS := '--go-grpc_opt=M$(USER_PB)' '--go-grpc_opt=M$(AUTH_PB)'

# `make tools` installs the protoc plugins here, which may not be on PATH.
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.DEFAULT_GOAL := help
.PHONY: help tools proto graphql generate run

help: ## List the targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-9s %s\n", $$1, $$2}'

tools: ## Install the protoc plugins for Go (gqlgen comes with go.mod)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: ## Regenerate the gRPC clients in pkg/pb from the .proto files
	rm -rf pkg/pb && mkdir -p pkg/pb
	protoc -I $(PROTO) \
		--go_out=pkg/pb      --go_opt=paths=source_relative      $(GO_OPTS) \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative $(GRPC_OPTS) \
		$(PROTOS)

graphql: ## Regenerate the GraphQL code in internal/graph from schema/
	go tool gqlgen generate

generate: proto graphql ## Regenerate everything

run: ## Start the gateway on :4000 (user-service :50051, auth-service :50052)
	go run ./cmd
