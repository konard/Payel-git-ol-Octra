FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Минимальный workspace: только agents (внешний gRPC-клиент) и сам orchestrator
RUN printf 'go 1.25.4\n\nuse (\n\t./agents\n\t./orchestrator\n)\n' > go.work

# Копируем agents — единственная внешняя gRPC-зависимость
COPY agents/go.mod agents/go.sum ./agents/
COPY agents/internal ./agents/internal
COPY agents/pkg ./agents/pkg
COPY agents/cmd ./agents/cmd
COPY agents/proto ./agents/proto
COPY agents/pb ./agents/pb

# Копируем orchestrator (boss + manager + worker в одном процессе)
COPY orchestrator/go.mod ./orchestrator/
COPY orchestrator/internal ./orchestrator/internal
COPY orchestrator/pkg ./orchestrator/pkg
COPY orchestrator/cmd ./orchestrator/cmd
COPY orchestrator/proto ./orchestrator/proto

RUN cd agents && go mod download
RUN cd orchestrator && go mod download

RUN cd orchestrator && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/orchestrator-bin ./cmd/app && ls -la /workspace/orchestrator-bin

FROM nixos/nix:latest

# git, ca-certificates, curl включены в образ nixos/nix

ENV NIX_CONFIG="experimental-features = nix-command flakes"
RUN nix-store --version && echo "nix is ready"

RUN mkdir -p /var/lib/octra/projects

COPY --from=builder /workspace/orchestrator-bin /usr/local/bin/orchestrator

RUN ls -la /usr/local/bin/ && \
    chmod +x /usr/local/bin/orchestrator

RUN printf '%s\n' \
    '#!/bin/sh' \
    'set -e' \
    '' \
    '# Start nix daemon if available (multi-user image), otherwise single-user mode' \
    'if command -v nix-daemon >/dev/null 2>&1; then' \
    '    nix-daemon --daemon >/dev/null 2>&1 &' \
    'fi' \
    '' \
    '# Configure git at runtime (env vars may not exist at build time)' \
    'git config --global user.email "${GIT_USER_EMAIL:-orchestrator@crewai.local}"' \
    'git config --global user.name  "${GIT_USER_NAME:-CrewAI Orchestrator}"' \
    'git config --global credential.helper store' \
    '' \
    'exec /usr/local/bin/orchestrator' \
    > /usr/local/bin/octra-entrypoint.sh && \
    chmod +x /usr/local/bin/octra-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/octra-entrypoint.sh"]
