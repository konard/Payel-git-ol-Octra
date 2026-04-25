package services

import (
	"errors"
	"time"
	"user/pkg/database"
	"user/pkg/models"
	"user/pkg/requests"

	"github.com/google/uuid"
)

// CreateWorkflow creates a new workflow in the library
func CreateWorkflow(userID uuid.UUID, username string, req requests.CreateWorkflowRequest) (*models.Workflow, error) {
	workflow := models.Workflow{
		ID:          uuid.New(),
		UserID:      userID,
		AuthorName:  username,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		IsPublic:    req.IsPublic,
		Downloads:   0,
	}

	err := database.Db.Create(&workflow).Error
	if err != nil {
		return nil, errors.New("failed to create workflow: " + err.Error())
	}

	return &workflow, nil
}

// GetWorkflowByID retrieves a workflow by its ID
func GetWorkflowByID(workflowID uuid.UUID, userID uuid.UUID) (*models.Workflow, error) {
	var workflow models.Workflow
	err := database.Db.Where("id = ? AND (is_public = true OR user_id = ?)", workflowID, userID).First(&workflow).Error
	if err != nil {
		return nil, errors.New("workflow not found")
	}
	return &workflow, nil
}

// GetUserWorkflows retrieves all workflows for a specific user
func GetUserWorkflows(userID uuid.UUID) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := database.Db.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error
	if err != nil {
		return nil, errors.New("failed to fetch workflows: " + err.Error())
	}
	return workflows, nil
}

// GetPublicWorkflows retrieves all public workflows (for the library)
func GetPublicWorkflows(category string, tag string, limit int, offset int) ([]models.Workflow, error) {
	var workflows []models.Workflow
	query := database.Db.Where("is_public = true")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if tag != "" {
		query = query.Where("tags @> ARRAY[?]::text[]", tag)
	}

	err := query.Order("downloads DESC").Limit(limit).Offset(offset).Find(&workflows).Error
	if err != nil {
		return nil, errors.New("failed to fetch public workflows: " + err.Error())
	}

	return workflows, nil
}

// UpdateWorkflow updates an existing workflow
func UpdateWorkflow(workflowID uuid.UUID, userID uuid.UUID, req requests.UpdateWorkflowRequest) (*models.Workflow, error) {
	var workflow models.Workflow
	err := database.Db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error
	if err != nil {
		return nil, errors.New("workflow not found or you don't have permission")
	}

	// Update fields
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	updates["is_public"] = req.IsPublic

	err = database.Db.Model(&workflow).Updates(updates).Error
	if err != nil {
		return nil, errors.New("failed to update workflow: " + err.Error())
	}

	// Reload the workflow
	err = database.Db.Where("id = ?", workflowID).First(&workflow).Error
	if err != nil {
		return nil, errors.New("failed to reload workflow: " + err.Error())
	}

	return &workflow, nil
}

// DeleteWorkflow deletes a workflow
func DeleteWorkflow(workflowID uuid.UUID, userID uuid.UUID) error {
	var workflow models.Workflow
	err := database.Db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error
	if err != nil {
		return errors.New("workflow not found or you don't have permission")
	}

	err = database.Db.Delete(&workflow).Error
	if err != nil {
		return errors.New("failed to delete workflow: " + err.Error())
	}

	return nil
}

// DownloadWorkflow increments the download counter for a workflow
func DownloadWorkflow(workflowID uuid.UUID) error {
	err := database.Db.Model(&models.Workflow{}).Where("id = ?", workflowID).UpdateColumn("downloads", database.Db.Raw("downloads + 1")).Error
	if err != nil {
		return errors.New("failed to update download count: " + err.Error())
	}
	return nil
}

// GetWorkflowCategories returns all unique workflow categories
func GetWorkflowCategories() ([]string, error) {
	var categories []string
	err := database.Db.Model(&models.Workflow{}).Distinct("category").Where("is_public = true AND category != ''").Pluck("category", &categories).Error
	if err != nil {
		return nil, errors.New("failed to fetch categories: " + err.Error())
	}
	return categories, nil
}
