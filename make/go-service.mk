# Shared targets for the Go services. A service's Makefile sets SERVICE, PORT
# and PROTOS, then includes this file:
#
#   SERVICE := user-service
#   PORT    := 50051
#   PROTOS  := user/user.proto user/events.proto auth/events.proto
#   include ../../make/go-service.mk
#
# Run `make` in a service folder to list the targets.

ROOT   := ../..
MODULE := github.com/jochem11/inventory-manager/services/$(SERVICE)
PROTO  := $(ROOT)/proto

# Settings and credentials (DB_PASSWORD, …) from the repo's .env, for run and
# test; see .env.example. Missing .env: the services' defaults apply.
-include $(ROOT)/.env
export

# Each .proto gets its Go package from its folder: user/user.proto becomes
# $(MODULE)/pkg/pb/user, package userpb. So the .proto files need no
# Go-specific go_package option.
pb_dir     = $(patsubst %/,%,$(dir $(1)))
pb_mapping = $(1)=$(MODULE)/pkg/pb/$(call pb_dir,$(1));$(call pb_dir,$(1))pb
GO_OPTS   := $(foreach p,$(PROTOS),'--go_opt=M$(call pb_mapping,$(p))')
GRPC_OPTS := $(foreach p,$(PROTOS),'--go-grpc_opt=M$(call pb_mapping,$(p))')

# `make tools` installs the protoc plugins here, which may not be on PATH.
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.DEFAULT_GOAL := help
.PHONY: help tools proto run test

help: ## List the targets
	@echo "$(SERVICE) (port $(PORT))"
	@grep -hE '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-9s %s\n", $$1, $$2}'

tools: ## Install the protoc plugins for Go
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: ## Regenerate pkg/pb from the .proto files
ifeq ($(strip $(PROTOS)),)
	@echo "$(SERVICE) has no .proto files yet"
else
	rm -rf pkg/pb && mkdir -p pkg/pb
	protoc -I $(PROTO) \
		--go_out=pkg/pb      --go_opt=paths=source_relative      $(GO_OPTS) \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative $(GRPC_OPTS) \
		$(PROTOS)
endif

run: ## Start the service on your machine (reads ../../.env)
	go run ./cmd

test: ## Run the tests; MySQL tests use the DB_* settings and skip without MySQL
	go test ./...
