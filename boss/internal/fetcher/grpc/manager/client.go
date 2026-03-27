package managerpb

import (
	"boss/internal/fetcher/grpc/managerpb"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client managerpb.ManagerServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: managerpb.NewManagerServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateManagers(ctx context.Context, taskID, title, description string, roles []string) (*managerpb.CreateManagersResponse, error) {
	req := &managerpb.CreateManagersRequest{
		TaskId:           taskID,
		Title:            title,
		Description:      description,
		Roles:            roles,
		QuantityManagers: int32(len(roles)),
	}

	resp, err := c.client.CreateManagers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("rpc failed: %w", err)
	}

	return resp, nil
}
