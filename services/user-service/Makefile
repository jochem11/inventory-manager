# user-service: run `make` to list the targets.

SERVICE := user-service
PORT    := 50051
# Its own contract and events, plus the auth events it consumes.
PROTOS  := user/user.proto user/events.proto auth/events.proto

include ../../make/go-service.mk
