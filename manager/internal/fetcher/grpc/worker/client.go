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
	if address == "" {
		address = "worker:50053"
	}
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
func (c *Client) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	resp, err := c.client.AssignWorkersAndWait(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("assign workers rpc failed: %w", err)
	}

	return resp, nil
}

// ReviewWorker отправляет замечания воркеру
func (c *Client) ReviewWorker(ctx context.Context, req *workerpb.ReviewRequest) (*workerpb.ReviewResponse, error) {
	resp, err := c.client.ReviewWorker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("review worker rpc failed: %w", err)
	}

	return resp, nil
}

// BuildAssignWorkersRequest helper для создания запроса
func BuildAssignWorkersRequest(taskID, managerID, managerRole, taskMD string,
	workerRoles []struct{ Role, Description string },
	tokens map[string]string, model, provider string,
	otherResults []*workerpb.WorkerResult) *workerpb.AssignWorkersRequest {

	tokensJSON, _ := json.Marshal(tokens)
	metadata := map[string]string{
		"tokens":   string(tokensJSON),
		"model":    model,
		"provider": provider,
	}
	// Also put provider-specific key
	if apiKey, ok := tokens[provider]; ok {
		metadata[provider] = apiKey
	}

	roles := make([]*workerpb.WorkerRole, len(workerRoles))
	for i, r := range workerRoles {
		roles[i] = &workerpb.WorkerRole{
			Role:        r.Role,
			Description: r.Description,
		}
	}

	return &workerpb.AssignWorkersRequest{
		TaskId:              taskID,
		ManagerId:           managerID,
		ManagerRole:         managerRole,
		WorkerRoles:         roles,
		TaskMd:              taskMD,
		Metadata:            metadata,
		OtherWorkersResults: otherResults,
	}
}
