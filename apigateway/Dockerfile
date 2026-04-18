FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Copy source code
COPY . .

RUN go mod tidy

# Собираем binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
