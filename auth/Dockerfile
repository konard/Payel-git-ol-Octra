FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY auth/go.mod auth/go.sum ./
COPY auth .
RUN go mod download && go mod tidy

# Собираем binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /auth-service ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 3112

CMD ["./auth-service"]
