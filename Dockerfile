# Build stage
FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /sigrok-mcp-server ./cmd/sigrok-mcp-server

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends sigrok-cli && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /sigrok-mcp-server /usr/local/bin/sigrok-mcp-server

ENTRYPOINT ["sigrok-mcp-server"]
