# Step 1: Standard Go Builder Image
FROM golang:1.25 AS builder
WORKDIR /workspace

# Copy the freshly cleaned structural module tracking maps
COPY . .

# Force the compiler to use local vendor files and ignore modern strict verification
ENV GOTOOLCHAIN=local
ENV GOFLAGS="-mod=vendor"

# Compile the monolithic OCM agent binary directly via standard native tools
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o resource-usage-collect-addon ./pkg

# Step 2: Minimal Distroless Execution Layer
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/resource-usage-collect-addon .
USER 65532:65532

ENTRYPOINT ["/resource-usage-collect-addon"]