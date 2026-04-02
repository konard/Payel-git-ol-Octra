package boss

import (
	"context"
	"fmt"

	"apigateway/internal/fetcher/grpc/boss/bosspb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client — gRPC клиент для связи с boss сервисом
type Client struct {
	conn   *grpc.ClientConn
	client bosspb.BossServiceClient
}

// NewClient подключается к boss сервису
func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to boss service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: bosspb.NewBossServiceClient(conn),
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// CreateTask отправляет задачу в boss сервис
func (c *Client) CreateTask(ctx context.Context, userID, username, title, description string, tokens map[string]string, meta map[string]string) (*bosspb.BossDecision, error) {
	req := &bosspb.CreateTaskRequest{
		UserId:      userID,
		Username:    username,
		Title:       title,
		Description: description,
		Tokens:      tokens,
		Meta:        meta,
	}

	resp, err := c.client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create task rpc failed: %w", err)
	}

	return resp, nil
}

// GetTaskStatus запрашивает статус задачи
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*bosspb.TaskStatusResponse, error) {
	req := &bosspb.TaskStatusRequest{
		TaskId: taskID,
	}

	resp, err := c.client.GetTaskStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get task status rpc failed: %w", err)
	}

	return resp, nil
}

// CreateTaskStream отправляет задачу в boss сервис и получает stream обновлений
func (c *Client) CreateTaskStream(ctx context.Context, req *bosspb.CreateTaskRequest) (bosspb.BossService_CreateTaskStreamClient, error) {
	stream, err := c.client.CreateTaskStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create task stream rpc failed: %w", err)
	}

	return stream, nil
}
