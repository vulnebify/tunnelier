# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tunnelier ./cmd/tunnelier

# Stage 2: Runtime
FROM alpine:latest

# Install wireguard + deps
RUN apk add --no-cache \
    ca-certificates \
    wireguard-tools \
    iproute2 \
    iptables \
    libc6-compat \
    libnss

WORKDIR /app

COPY --from=builder /app/tunnelier .

ENTRYPOINT ["./tunnelier"]
