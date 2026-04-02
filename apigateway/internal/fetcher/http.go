package fetcher

import (
	"context"
	"crewai/internal/fetcher/grpc/boss"
	"crewai/pkg/requests"
	"log"

	"github.com/Payel-git-ol/azure"
)

var bossClient *boss.Client

func init() {
	var err error
	bossHost := "boss:50051"
	bossClient, err = boss.NewClient(bossHost)
	if err != nil {
		log.Printf("Warning: failed to connect to boss service: %v", err)
	}
}

func HttpManager(a *azure.Azure) {
	a.Use(azure.Recovery())
	a.Use(azure.Logger())

	a.Get("/health", func(c *azure.Context) {
		c.JsonStatus(200, azure.M{
			"status": "OK",
		})
	})

	a.Post("/task/create", func(c *azure.Context) {
		var req requests.CreateTaskRequest

		if err := c.BindJSON(&req); err != nil {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		if bossClient == nil {
			c.JsonStatus(503, azure.M{
				"status": "error",
				"error":  "boss service unavailable",
			})
			return
		}

		// Call Boss and wait for full cycle
		resp, err := bossClient.CreateTask(context.Background(), req.UserID, req.Username, req.Title, req.Description, req.Tokens, req.Meta)
		if err != nil {
			c.JsonStatus(500, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		// Check status
		if resp.Status == "error" {
			c.JsonStatus(500, azure.M{
				"status":  "error",
				"message": resp.ErrorMessage,
			})
			return
		}

		// Return ZIP archive to user
		c.JsonStatus(200, azure.M{
			"status":       "success",
			"task_id":      resp.TaskId,
			"managers":     resp.ManagersCount,
			"tech_stack":   resp.TechStack,
			"description":  resp.TechnicalDescription,
			"architecture": resp.ArchitectureNotes,
			"solution":     resp.Solution, // ZIP archive bytes
		})
	})

	a.Get("/task/status", func(c *azure.Context) {
		taskID := c.GetQueryParam("task_id")
		if taskID == "" {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "task_id is required",
			})
			return
		}

		if bossClient == nil {
			c.JsonStatus(503, azure.M{
				"status": "error",
				"error":  "boss service unavailable",
			})
			return
		}

		resp, err := bossClient.GetTaskStatus(context.Background(), taskID)
		if err != nil {
			c.JsonStatus(500, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JsonStatus(200, azure.M{
			"status":   "success",
			"task_id":  resp.TaskId,
			"progress": resp.Progress,
		})
	})
}
