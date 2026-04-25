package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"user/pkg/database"
	"user/pkg/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" || os.Getenv("REDIS_ENABLED") != "true" {
		return
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Printf("Failed to parse Redis URL: %v\n", err)
		return
	}

	redisClient = redis.NewClient(opt)
}

func GetChatHistory(userID uuid.UUID) ([]models.Chat, error) {
	var cacheKey string
	if redisClient != nil {
		cacheKey = fmt.Sprintf("chat:history:%s", userID.String())
		cached, err := redisClient.Get(nil, cacheKey).Result()
		if err == nil {
			var chats []models.Chat
			if json.Unmarshal([]byte(cached), &chats) == nil {
				return chats, nil
			}
		}
	}

	var chats []models.Chat
	err := database.Db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(50).
		Find(&chats).Error
	if err != nil {
		return nil, err
	}

	if redisClient != nil && cacheKey != "" {
		if data, err := json.Marshal(chats); err == nil {
			redisClient.Set(nil, cacheKey, data, 5*time.Minute)
		}
	}

	return chats, nil
}

func CreateChat(userID uuid.UUID, title string) (*models.Chat, error) {
	if title == "" {
		title = "New Chat"
	}

	chat := &models.Chat{
		Id:          uuid.New(),
		UserId:       userID,
		Title:       title,
		LastMessage: "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := database.Db.Create(chat).Error
	if err != nil {
		return nil, err
	}

	invalidateChatHistoryCache(userID)

	return chat, nil
}

func GetChat(userID uuid.UUID, chatID string) (*models.Chat, error) {
	id, err := uuid.Parse(chatID)
	if err != nil {
		return nil, errors.New("invalid chat ID")
	}

	if redisClient != nil {
		cacheKey := fmt.Sprintf("chat:%s", chatID)
		cached, err := redisClient.Get(nil, cacheKey).Result()
		if err == nil {
			var chat models.Chat
			if json.Unmarshal([]byte(cached), &chat) == nil && chat.UserId == userID {
				database.Db.Where("chat_id = ?", chatID).Order("created_at ASC").Find(&chat.Messages)
				return &chat, nil
			}
		}
	}

	var chat models.Chat
	err = database.Db.Where("id = ? AND user_id = ?", id, userID).First(&chat).Error
	if err != nil {
		return nil, err
	}

	database.Db.Where("chat_id = ?", chatID).Order("created_at ASC").Find(&chat.Messages)

	if redisClient != nil {
		if data, err := json.Marshal(chat); err == nil {
			cacheKey := fmt.Sprintf("chat:%s", chatID)
			redisClient.Set(nil, cacheKey, data, 5*time.Minute)
		}
	}

	return &chat, nil
}

func AddChatMessage(userID uuid.UUID, chatID, role, content, provider, model, apiKeyID string) (*models.ChatMessage, error) {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return nil, errors.New("invalid chat ID")
	}

	var chat models.Chat
	err = database.Db.Where("id = ? AND user_id = ?", chatUUID, userID).First(&chat).Error
	if err != nil {
		return nil, errors.New("chat not found")
	}

	msg := &models.ChatMessage{
		Id:        uuid.New(),
		ChatId:    chatUUID,
		Role:      role,
		Content:   content,
		Provider:  provider,
		Model:     model,
		ApiKeyId:  apiKeyID,
		CreatedAt: time.Now(),
	}

	err = database.Db.Create(msg).Error
	if err != nil {
		return nil, err
	}

	chat.LastMessage = content
	chat.UpdatedAt = time.Now()

	if chat.Provider == "" && provider != "" {
		chat.Provider = provider
	}
	if chat.Model == "" && model != "" {
		chat.Model = model
	}

	database.Db.Save(&chat)

	invalidateChatCache(chatID)
	invalidateChatHistoryCache(userID)

	return msg, nil
}

func UpdateChatTitle(chatID, title string) error {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return errors.New("invalid chat ID")
	}

	result := database.Db.Model(&models.Chat{}).
		Where("id = ?", chatUUID).
		Updates(map[string]interface{}{"title": title, "updated_at": time.Now()})

	if result.Error != nil {
		return result.Error
	}

	invalidateChatCache(chatID)
	return nil
}

func UpdateChatWorkflow(chatID, workflow string) error {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return errors.New("invalid chat ID")
	}

	result := database.Db.Model(&models.Chat{}).
		Where("id = ?", chatUUID).
		Updates(map[string]interface{}{"workflow": workflow, "updated_at": time.Now()})

	if result.Error != nil {
		return result.Error
	}

	invalidateChatCache(chatID)
	return nil
}

func DeleteChat(chatID string) error {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return errors.New("invalid chat ID")
	}

	database.Db.Where("chat_id = ?", chatUUID).Delete(&models.ChatMessage{})

	result := database.Db.Delete(&models.Chat{}, "id = ?", chatUUID).Error
	if result != nil {
		return result
	}

	invalidateChatCache(chatID)
	return nil
}

func invalidateChatCache(chatID string) {
	if redisClient == nil {
		return
	}
	redisClient.Del(nil, fmt.Sprintf("chat:%s", chatID))
}

func invalidateChatHistoryCache(userID uuid.UUID) {
	if redisClient == nil {
		return
	}
	redisClient.Del(nil, fmt.Sprintf("chat:history:%s", userID.String()))
}