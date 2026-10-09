# auth-service: run `make` to list the targets.

SERVICE := auth-service
PORT    := 50052
# Its own contract and events, plus the user events it consumes.
PROTOS  := auth/auth.proto auth/events.proto user/events.proto

include ../../make/go-service.mk
