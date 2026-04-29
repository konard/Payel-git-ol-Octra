FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Создаём минимальный go.work
RUN printf 'go 1.25.4\n\nuse (\n\t./agents\n\t./worker\n)\n' > go.work

# Копируем агентов
COPY agents/go.mod agents/go.sum ./agents/
COPY agents/internal ./agents/internal
COPY agents/pkg ./agents/pkg
COPY agents/cmd ./agents/cmd
COPY agents/proto ./agents/proto
COPY agents/pb ./agents/pb

# Копируем worker
COPY worker/go.mod worker/go.sum ./worker/
COPY worker/internal ./worker/internal
COPY worker/pkg ./worker/pkg
COPY worker/cmd ./worker/cmd
COPY worker/proto ./worker/proto

RUN cd agents && go mod download
RUN cd worker && go mod download

RUN cd worker && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates git

COPY --from=builder /workspace/main .

CMD ["./main"]
