FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Минимальный workspace: только agents (внешний gRPC-клиент) и сам nodes
RUN printf 'go 1.25.4\n\nuse (\n\t./agents\n\t./nodes\n)\n' > go.work

# Копируем agents — единственная внешняя gRPC-зависимость
COPY agents/go.mod agents/go.sum ./agents/
COPY agents/internal ./agents/internal
COPY agents/pkg ./agents/pkg
COPY agents/cmd ./agents/cmd
COPY agents/proto ./agents/proto
COPY agents/pb ./agents/pb

# Копируем nodes (boss + manager + worker в одном процессе)
COPY nodes/go.mod nodes/go.sum ./nodes/
COPY nodes/internal ./nodes/internal
COPY nodes/pkg ./nodes/pkg
COPY nodes/cmd ./nodes/cmd
COPY nodes/proto ./nodes/proto

RUN cd agents && go mod download
RUN cd nodes && go mod download

RUN cd nodes && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/nodes ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates git

RUN git config --global user.email "${GIT_USER_EMAIL:-nodes@crewai.local}" && \
    git config --global user.name "${GIT_USER_NAME:-CrewAI Nodes}" && \
    git config --global credential.helper store

COPY --from=builder /workspace/nodes .

CMD ["./nodes"]
