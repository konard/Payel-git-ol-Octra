FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Копируем все модули и go.work
COPY go.work ./
COPY agents ./agents
COPY boss ./boss
COPY manager ./manager
COPY worker ./worker

# Скачиваем зависимости для agents
RUN cd agents && go mod download

# Собираем binary
RUN cd agents && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

# Финальный образ
FROM alpine:latest

WORKDIR /root/

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копируем бинарник
COPY --from=builder /workspace/main .

# Запускаем
CMD ["./main"]
