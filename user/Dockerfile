FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY user/go.mod user/go.sum ./
COPY user/ .
RUN go mod download && go mod tidy

# Собираем binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /user-service ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /user-service .

EXPOSE 3112

CMD ["./user-service"]
