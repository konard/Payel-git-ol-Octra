package main

import (
	"agents/internal/fetcher/providers/claude"
	"agents/internal/fetcher/providers/deepseek"
	"agents/internal/fetcher/providers/gemini"
	"agents/internal/fetcher/providers/grok"
	"agents/internal/fetcher/providers/openai"
	"agents/internal/fetcher/providers/openrouter"
	"agents/internal/fetcher/providers/qwen"
	"agents/internal/service"
	"agents/pkg/models"
)

func InitProvider(providers map[string]*models.ProviderConfig, agentService *service.AgentService) {
	if p, err := claude.New(providers["claude"]); err == nil {
		agentService.RegisterProvider("claude", p)
	}
	if p, err := gemini.New(providers["gemini"]); err == nil {
		agentService.RegisterProvider("gemini", p)
	}
	if p, err := openai.New(providers["openai"]); err == nil {
		agentService.RegisterProvider("openai", p)
	}
	if p, err := deepseek.New(providers["deepseek"]); err == nil {
		agentService.RegisterProvider("deepseek", p)
	}
	if p, err := grok.New(providers["grok"]); err == nil {
		agentService.RegisterProvider("grok", p)
	}
	if p, err := openrouter.New(providers["openrouter"]); err == nil {
		agentService.RegisterProvider("openrouter", p)
	}
	if p, err := qwen.New(providers["qwen"]); err == nil {
		agentService.RegisterProvider("cliproxy", p)
	}
}
