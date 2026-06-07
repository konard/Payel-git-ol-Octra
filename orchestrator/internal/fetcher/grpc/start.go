package grpc

import (
	"log"
	"net"

	"orchestrator/internal/fetcher/grpc/bosspb"
	"orchestrator/internal/redis"
	"orchestrator/internal/service/rules/boss"

	"google.golang.org/grpc"
)

// Start — запускает внешний gRPC-сервер orchestrator для apigateway.
// Внутри orchestrator больше нет инсервисных gRPC-вызовов — только этот.
func Start(port string, b *boss.Service, r *redis.Client) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)
	bosspb.RegisterBossServiceServer(grpcServer, NewServer(b, r))
	log.Printf("orchestrator gRPC server listening on port %s", port)
	return grpcServer.Serve(lis)
}

