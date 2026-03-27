package fetcher

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/pkg/models"
	"context"
	"github.com/yhwhpe/llm-unified-client"
	"log"
	"os"
)

type HttpManagerRepository interface {
	SendTaskWithAiBoss(req *bosspb.TaskRequest)
}

type HttpManager struct{}

func (s *HttpManager) SendTaskWithAiBoss(req *bosspb.TaskRequest) {
	config := llm.Config{
		Provider:     llm.ProviderDeepSeek,
		APIKey:       "your-api-key",
		BaseURL:      "https://api.deepseek.com",
		DefaultModel: "deepseek-chat",
	}

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	promt, err := os.ReadFile("PROMT.md")
	if err != nil {
		log.Printf("Failed to read PROMT.md file: %v", err)
	}

	task := models.Task{
		TaskName:    req.Taskname,
		Title:       req.Title,
		Description: req.Description,
	}

	mess := "Promt: " + string(promt) + "Title task: " + task.Title + "Description task: " + task.Description

	response, err := llm.GenerateSimple(context.Background(), client, mess)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response.Content)
}
