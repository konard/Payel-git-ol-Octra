FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git

COPY agents/go.mod agents/go.sum ./agents/
COPY agents/internal ./agents/internal
COPY agents/pkg ./agents/pkg
COPY agents/cmd ./agents/cmd
COPY agents/proto ./agents/proto
COPY agents/pb ./agents/pb

RUN cd agents && go mod download && go mod tidy

RUN cd agents && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main ./cmd/app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/main .

CMD ["./main"]
