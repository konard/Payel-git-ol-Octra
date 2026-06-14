package grpc

import (
	"context"
	"fmt"

	"user/internal/grpc/bosspb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrchestratorClient struct {
	conn   *grpc.ClientConn
	client bosspb.BossServiceClient
}

func NewOrchestratorClient(address string) (*OrchestratorClient, error) {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(500*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	return &OrchestratorClient{
		conn:   conn,
		client: bosspb.NewBossServiceClient(conn),
	}, nil
}

func (c *OrchestratorClient) RestoreProjectFiles(ctx context.Context, nixStorePath, taskID string) (*bosspb.RestoreProjectFilesResponse, error) {
	return c.client.RestoreProjectFiles(ctx, &bosspb.RestoreProjectFilesRequest{
		NixStorePath: nixStorePath,
		TaskId:       taskID,
	})
}

func (c *OrchestratorClient) Close() error {
	return c.conn.Close()
}
