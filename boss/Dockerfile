FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости
RUN apk add --no-cache git

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# Финальный образ
FROM alpine:latest

WORKDIR /root/

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копируем бинарник
COPY --from=builder /app/main .

# Запускаем
CMD ["./main"]
