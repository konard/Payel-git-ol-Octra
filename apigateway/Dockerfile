FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Копируем apigateway и boss модули (apigateway зависит от boss)
COPY apigateway ./apigateway
COPY boss ./boss

RUN cd apigateway && go mod download

RUN cd apigateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
