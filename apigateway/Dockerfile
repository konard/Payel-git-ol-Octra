FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Apigateway is self-contained (proto files are copied locally)
COPY apigateway ./apigateway

RUN cd apigateway && go mod tidy

# Собираем binary
RUN cd apigateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
