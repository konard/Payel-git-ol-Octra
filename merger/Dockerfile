FROM golang:1.21-alpine AS builder

WORKDIR /app/merger

COPY merger/go.mod merger/go.sum ./
RUN go mod download

COPY merger/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /merger ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /merger .

EXPOSE 50054

CMD ["./merger"]
