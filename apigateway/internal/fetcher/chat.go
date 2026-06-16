package fetcher

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func handleBossChatWS(conn *websocket.Conn, streamID, userID, username string, initial []byte) {
	defer conn.Close()

	done := make(chan struct{})
	go PingWriter(conn, done)
	defer close(done)

	handleFrame := func(data []byte) bool {
		if len(bytes.TrimSpace(data)) == 0 {
			return false
		}

		var msg chatWSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Invalid chat JSON: " + err.Error(),
			})
			return false
		}

		if msg.Type == "ping" {
			conn.WriteJSON(gin.H{"type": "pong"})
			return false
		}
		if msg.Type != "chat" {
			return false
		}

		message := strings.TrimSpace(msg.Message)
		if message == "" {
			return false
		}

		if shouldLaunchWorkflowFromChat(message) {
			if msg.TaskPayload == nil {
				writeBossChatMessage(conn, streamID, "I can start a workflow from chat, but I need your model settings first.", true)
				return false
			}

			taskReq := *msg.TaskPayload
			taskReq.UserID = userID
			taskReq.Username = username
			if taskReq.Title == "" {
				taskReq.Title = message
			}
			if taskReq.Description == "" {
				taskReq.Description = message
			}
			if taskReq.Meta == nil {
				taskReq.Meta = map[string]interface{}{}
			}

			writeBossChatMessage(conn, streamID, "I'll start the workflow for this request.", false)
			processTaskStreamWS(conn, taskReq, streamID)
			return true
		}

		writeBossChatMessage(conn, streamID, buildBossChatReply(message), false)
		return false
	}

	if handleFrame(initial) {
		return
	}

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		switch msgType {
		case websocket.CloseMessage:
			return
		case websocket.PingMessage:
			conn.WriteMessage(websocket.PongMessage, nil)
		case websocket.TextMessage:
			if handleFrame(data) {
				return
			}
		}
	}
}

func writeBossChatMessage(conn *websocket.Conn, taskID, message string, clarification bool) {
	conn.WriteJSON(gin.H{
		"type":             "chat",
		"task_id":          taskID,
		"sender":           "boss",
		"message":          message,
		"is_clarification": clarification,
		"timestamp":        time.Now().Unix(),
	})
}

func containsGitHubTaskURL(message string) bool {
	return githubTaskURLPattern.MatchString(message)
}

func shouldLaunchWorkflowFromChat(message string) bool {
	if containsGitHubTaskURL(message) {
		return true
	}

	words := normalizedWords(message)
	if shouldLaunchSearchWorkflowFromChat(words) {
		return true
	}

	triggers := []string{
		"build", "building", "create", "creating", "develop", "fix", "generate", "generating",
		"implement", "implementing", "launch", "make", "refactor", "run", "scaffold", "start", "write",
		"введи", "добавь", "запусти", "исправь", "напиши", "подготовь", "разработай",
		"сгенерируй", "сделай", "создай", "создать", "собери",
	}
	targets := []string{
		"app", "application", "api", "backend", "bug", "code", "component", "feature", "frontend", "proxy",
		"integration", "page", "project", "service", "site", "tool", "webapp", "website", "workflow",
		"бот", "доклад", "документ", "код", "презентацией", "презентации", "презентацию",
		"презентация", "приложение", "проект", "прокси", "сайт", "сервис", "слайд", "слайдами",
		"слайды", "функцию",
	}
	return hasAnyWord(words, triggers) && hasAnyWord(words, targets)
}

func shouldLaunchSearchWorkflowFromChat(words map[string]bool) bool {
	searchTriggers := []string{
		"find", "google", "lookup", "research", "search",
		"найди", "погугли", "поиск", "поищи",
	}
	if hasAnyWord(words, searchTriggers) || (words["look"] && words["up"]) {
		return true
	}

	searchTargets := []string{
		"configure", "docs", "documentation", "install", "latest", "links", "news", "reference", "setup",
		"документация", "документацию", "настроить", "новости", "ссылки", "установить",
	}
	if hasAnyWord(words, []string{"how", "как"}) && hasAnyWord(words, searchTargets) {
		return true
	}

	if len(words) > 1 && hasAnyWord(words, searchTargets) {
		return true
	}

	return false
}

func buildBossChatReply(message string) string {
	words := normalizedWords(message)
	if hasAnyWord(words, []string{"hello", "hi", "hey"}) {
		return "Hi, I'm here."
	}
	if hasAnyWord(words, []string{"thanks", "thank"}) {
		return "You're welcome."
	}
	if hasAnyWord(words, []string{"help", "workflow", "build", "create"}) {
		return "I can answer here, or start the workflow when you describe code you want built or changed."
	}
	return "I understand. Send the next detail when you are ready."
}

func normalizedWords(message string) map[string]bool {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, message)

	words := make(map[string]bool)
	for _, word := range strings.Fields(cleaned) {
		words[word] = true
	}
	return words
}

func hasAnyWord(words map[string]bool, candidates []string) bool {
	for _, candidate := range candidates {
		if words[candidate] {
			return true
		}
	}
	return false
}
