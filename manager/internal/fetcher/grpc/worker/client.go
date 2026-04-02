package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"manager/internal/fetcher/grpc/workerpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client — gRPC клиент для связи с Worker сервисом
type Client struct {
	conn   *grpc.ClientConn
	client workerpb.WorkerServiceClient
}

// NewClient подключается к Worker сервису
func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: workerpb.NewWorkerServiceClient(conn),
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// AssignWorkersAndWait отправляет запрос и ждёт ZIP архив
func (c *Client) AssignWorkersAndWait(ctx context.Context, taskID, managerID, managerRole, taskMD string, workerRoles, tokens []string, model, modelURL string) ([]byte, error) {
	tokensJSON, _ := json.Marshal(tokens)
	metadata := map[string]string{
		"tokens":   string(tokensJSON),
		"model":    model,
		"modelUrl": modelURL,
	}

	roles := make([]*workerpb.WorkerRole, len(workerRoles))
	for i, r := range workerRoles {
		roles[i] = &workerpb.WorkerRole{
			Role:        r,
			Description: "Worker role",
		}
	}

	req := &workerpb.AssignWorkersRequest{
		TaskId:      taskID,
		ManagerId:   managerID,
		ManagerRole: managerRole,
		WorkerRoles: roles,
		TaskMd:      taskMD,
		Metadata:    metadata,
	}

	resp, err := c.client.AssignWorkersAndWait(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("assign workers rpc failed: %w", err)
	}

	return resp.Solution, nil
}
