package services

import (
	"errors"
	"fmt"
	"user/pkg/models"
	"user/pkg/requests"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomProviderService struct {
	db *gorm.DB
}

func NewCustomProviderService(db *gorm.DB) *CustomProviderService {
	return &CustomProviderService{db: db}
}

// CreateCustomProvider creates a new custom provider for a user
func (s *CustomProviderService) CreateCustomProvider(userID uuid.UUID, req requests.CreateCustomProviderRequest) (*models.CustomProvider, error) {
	provider := &models.CustomProvider{
		UserID:         userID,
		Name:           req.Name,
		BaseURL:        req.BaseURL,
		APIKey:         req.APIKey,
		RequiresApiKey: req.RequiresApiKey,
	}

	if err := s.db.Create(provider).Error; err != nil {
		return nil, fmt.Errorf("failed to create custom provider: %w", err)
	}

	return provider, nil
}

// GetUserCustomProviders returns all custom providers for a user
func (s *CustomProviderService) GetUserCustomProviders(userID uuid.UUID) ([]models.CustomProvider, error) {
	var providers []models.CustomProvider
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&providers).Error; err != nil {
		return nil, fmt.Errorf("failed to get custom providers: %w", err)
	}
	return providers, nil
}

// UpdateCustomProvider updates a custom provider
func (s *CustomProviderService) UpdateCustomProvider(userID, providerID uuid.UUID, req requests.UpdateCustomProviderRequest) (*models.CustomProvider, error) {
	var provider models.CustomProvider
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", providerID, userID).First(&provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("custom provider not found")
		}
		return nil, fmt.Errorf("failed to find custom provider: %w", err)
	}

	// Update fields
	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.BaseURL != "" {
		provider.BaseURL = req.BaseURL
	}
	if req.APIKey != "" || (req.APIKey == "" && req.RequiresApiKey != nil && !*req.RequiresApiKey) {
		provider.APIKey = req.APIKey
	}
	if req.RequiresApiKey != nil {
		provider.RequiresApiKey = *req.RequiresApiKey
	}

	if err := s.db.Save(&provider).Error; err != nil {
		return nil, fmt.Errorf("failed to update custom provider: %w", err)
	}

	return &provider, nil
}

// DeleteCustomProvider soft deletes a custom provider
func (s *CustomProviderService) DeleteCustomProvider(userID, providerID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", providerID, userID).Delete(&models.CustomProvider{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete custom provider: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("custom provider not found")
	}
	return nil
}

// GetCustomProvider returns a specific custom provider
func (s *CustomProviderService) GetCustomProvider(userID, providerID uuid.UUID) (*models.CustomProvider, error) {
	var provider models.CustomProvider
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", providerID, userID).First(&provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("custom provider not found")
		}
		return nil, fmt.Errorf("failed to find custom provider: %w", err)
	}
	return &provider, nil
}

// CreateCustomModel creates a new custom model for a user
func (s *CustomProviderService) CreateCustomModel(userID uuid.UUID, req requests.CreateCustomModelRequest) (*models.CustomModel, error) {
	model := &models.CustomModel{
		UserID: userID,
		Name:   req.Name,
	}

	// Only set ProviderID if it's provided and not a zero UUID
	if req.ProviderID != nil && *req.ProviderID != uuid.Nil {
		model.ProviderID = req.ProviderID
	}

	if err := s.db.Create(model).Error; err != nil {
		return nil, fmt.Errorf("failed to create custom model: %w", err)
	}

	return model, nil
}

// GetUserCustomModels returns all custom models for a user
func (s *CustomProviderService) GetUserCustomModels(userID uuid.UUID) ([]models.CustomModel, error) {
	var models []models.CustomModel
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get custom models: %w", err)
	}
	return models, nil
}

// UpdateCustomModel updates a custom model
func (s *CustomProviderService) UpdateCustomModel(userID, modelID uuid.UUID, req requests.UpdateCustomModelRequest) (*models.CustomModel, error) {
	var model models.CustomModel
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", modelID, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("custom model not found")
		}
		return nil, fmt.Errorf("failed to find custom model: %w", err)
	}

	// Update fields
	if req.Name != "" {
		model.Name = req.Name
	}
	if req.ProviderID != nil && *req.ProviderID != uuid.Nil {
		model.ProviderID = req.ProviderID
	}

	if err := s.db.Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update custom model: %w", err)
	}

	return &model, nil
}

// DeleteCustomModel soft deletes a custom model
func (s *CustomProviderService) DeleteCustomModel(userID, modelID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", modelID, userID).Delete(&models.CustomModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete custom model: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("custom model not found")
	}
	return nil
}

// GetCustomModel returns a specific custom model
func (s *CustomProviderService) GetCustomModel(userID, modelID uuid.UUID) (*models.CustomModel, error) {
	var model models.CustomModel
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", modelID, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("custom model not found")
		}
		return nil, fmt.Errorf("failed to find custom model: %w", err)
	}
	return &model, nil
}
