# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
ENV GOPROXY=https://proxy.golang.org,direct
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /octra-backend ./cmd/server

# Runtime stage — uses the Nix image so the backend can create and manage
# per-user Nix environments. Falls back to plain shell when Nix is unavailable.
FROM nixos/nix:latest

# Enable flakes / nix-command, used when provisioning user environments.
RUN mkdir -p /etc/nix && \
    echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf

WORKDIR /app
COPY --from=builder /octra-backend /app/octra-backend

EXPOSE 8080

CMD ["/app/octra-backend"]
