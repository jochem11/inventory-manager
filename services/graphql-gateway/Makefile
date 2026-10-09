# graphql-gateway: run `make` to list the targets.

SERVICE := graphql-gateway
PORT    := 4000
# The gRPC clients of the services it calls.
PROTOS  := user/user.proto auth/auth.proto item/item.proto

include ../../make/go-service.mk

.PHONY: graphql generate

graphql: ## Regenerate the GraphQL code in internal/graph from schema/
	go tool gqlgen generate

generate: proto graphql ## Regenerate everything
