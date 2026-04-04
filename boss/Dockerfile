FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Устанавливаем зависимости
RUN apk add --no-cache git

# Копируем все модули и go.work
COPY go.work ./
COPY agents ./agents
COPY apigateway ./apigateway
COPY boss ./boss
COPY manager ./manager
COPY merger ./merger
COPY worker ./worker

# Скачиваем зависимости для boss
RUN cd boss && go mod download

# Собираем binary
RUN cd boss && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

# Финальный образ
FROM alpine:latest

WORKDIR /root/

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копируем бинарник
COPY --from=builder /workspace/main .

# Запускаем
CMD ["./main"]
