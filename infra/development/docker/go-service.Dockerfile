# Go services — development. Tilt compiles the binary on your machine (fast,
# with your Go build cache) into build/<service>; the image only copies it in.
# Build context: build/. BINARY: the service name, e.g. user-service.
FROM alpine:3.22
ARG BINARY
COPY ${BINARY} /usr/local/bin/service
ENTRYPOINT ["/usr/local/bin/service"]
