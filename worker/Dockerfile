FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

# Копируем все модули и go.work
COPY go.work ./
COPY agents ./agents
COPY boss ./boss
COPY manager ./manager
COPY worker ./worker

RUN cd worker && go mod download

RUN cd worker && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
