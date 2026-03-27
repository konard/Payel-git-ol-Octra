package manager

import (
	"context"
	"fmt"

	"boss/internal/fetcher/grpc/manager/managerspb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client — gRPC клиент для связи с manager сервисом
type Client struct {
	client managerspb.ManagerServiceClient
	conn   *grpc.ClientConn
}

// NewClient создаёт подключение к manager сервису
func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to manager service: %w", err)
	}

	return &Client{
		client: managerspb.NewManagerServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// CreateManagers отправляет запрос на создание команды менеджеров
func (c *Client) CreateManagers(ctx context.Context, taskID, taskname, title, description, username string, quantity int32, roles []string, metadata map[string]string) (*managerspb.CreateManagersResponse, error) {
	req := &managerspb.CreateManagersRequest{
		TaskId:           taskID,
		Taskname:         taskname,
		Title:            title,
		Description:      description,
		Username:         username,
		QuantityManagers: quantity,
		Roles:            roles,
		Metadata:         metadata,
	}

	resp, err := c.client.CreateManagers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create managers rpc failed: %w", err)
	}

	return resp, nil
}

// GetTaskStatus запрашивает статус задачи
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*managerspb.TaskStatusResponse, error) {
	req := &managerspb.TaskStatusRequest{
		TaskId: taskID,
	}

	resp, err := c.client.GetTaskStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get task status rpc failed: %w", err)
	}

	return resp, nil
}
