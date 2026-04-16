package manager

import (
	"context"
	"encoding/json"
	"fmt"

	"boss/internal/fetcher/grpc/manager/managerpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client — gRPC клиент для связи с Manager сервисом
type Client struct {
	conn   *grpc.ClientConn
	client managerpb.ManagerServiceClient
}

// NewClient подключается к Manager сервису
func NewClient(address string) (*Client, error) {
	if address == "" {
		address = "manager:50052"
	}
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to manager service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: managerpb.NewManagerServiceClient(conn),
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// AssignManager вызывает ОДНОГО менеджера (новый метод)
func (c *Client) AssignManager(ctx context.Context, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
	resp, err := c.client.AssignManager(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("assign manager rpc failed: %w", err)
	}

	return resp, nil
}

// AssignManagerStream вызывает ОДНОГО менеджера с progress updates
func (c *Client) AssignManagerStream(ctx context.Context, req *managerpb.AssignManagerRequest) (managerpb.ManagerService_AssignManagerStreamClient, error) {
	stream, err := c.client.AssignManagerStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("assign manager stream rpc failed: %w", err)
	}

	return stream, nil
}

// AssignManagersAndWait отправляет запрос и ждёт ZIP архив (legacy)
func (c *Client) AssignManagersAndWait(ctx context.Context, taskID, techDesc string, roles []string, tokens map[string]string, model, provider string) ([]byte, error) {
	tokensJSON, _ := json.Marshal(tokens)
	metadata := map[string]string{
		"tokens":   string(tokensJSON),
		"model":    model,
		"provider": provider,
	}

	req := &managerpb.AssignManagersRequest{
		TaskId:               taskID,
		TechnicalDescription: techDesc,
		Roles:                roles,
		Metadata:             metadata,
	}

	resp, err := c.client.AssignManagersAndWait(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("assign managers rpc failed: %w", err)
	}

	return resp.Solution, nil
}
