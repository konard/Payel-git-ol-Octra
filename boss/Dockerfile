FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Устанавливаем зависимости
RUN apk add --no-cache git

# Создаём go.work для boss с зависимостями
RUN printf 'go 1.25.4\n\nuse (\n\t./agents\n\t./boss\n\t./manager\n)\n' > go.work

# Копируем агентов целиком
COPY agents/go.mod agents/go.sum ./agents/
COPY agents/internal ./agents/internal
COPY agents/pkg ./agents/pkg
COPY agents/cmd ./agents/cmd
COPY agents/proto ./agents/proto
COPY agents/pb ./agents/pb

# Копируем manager целиком
COPY manager/go.mod manager/go.sum ./manager/
COPY manager/internal ./manager/internal
COPY manager/pkg ./manager/pkg
COPY manager/cmd ./manager/cmd
COPY manager/proto ./manager/proto

# Копируем boss целиком
COPY boss/go.mod boss/go.sum ./boss/
COPY boss/internal ./boss/internal
COPY boss/pkg ./boss/pkg
COPY boss/cmd ./boss/cmd
COPY boss/proto ./boss/proto

# Скачиваем зависимости напрямую
RUN cd agents && go mod download
RUN cd manager && go mod download
RUN cd boss && go mod download

# Собираем binary
RUN cd boss && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
