FROM golang:1.21 AS builder
WORKDIR /go/src/cluster-info-addon
COPY . .
ENV GO_PACKAGE cluster-info-addon

# Build
RUN make build --warn-undefined-variables

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

# Add the binaries
COPY --from=builder /go/src/cluster-info-addon/bin/cluster-info-addon /

ENTRYPOINT ["/cluster-info-addon"]
