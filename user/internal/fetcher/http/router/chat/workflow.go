package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/pkg/requests"
)

func parseWorkflowID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid workflow id"})
		return uuid.Nil, false
	}
	return id, true
}

func workflowQueryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func ChatIdWorkflowPut(c *gin.Context) {

}

func WorkflowsPost(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	username := c.GetString("username")
	if username == "" {
		username = "user"
	}

	var req requests.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid request body"})
		return
	}

	workflow, err := services.CreateWorkflow(userID, username, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": workflow})
}

func WorkflowsLibrary(c *gin.Context) {
	limit := workflowQueryInt(c, "limit", 50)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := workflowQueryInt(c, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	workflows, err := services.GetPublicWorkflows(c.Query("category"), c.Query("tag"), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": workflows})
}

func WorkflowsCategories(c *gin.Context) {
	categories, err := services.GetWorkflowCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": categories})
}

func WorkflowsMy(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	workflows, err := services.GetUserWorkflows(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": workflows})
}

func GetWorkflowsId(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok {
		return
	}

	workflow, err := services.GetWorkflowByID(workflowID, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": workflow})
}

func WorkflowsIdDownloadPost(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok {
		return
	}
	if err := services.DownloadWorkflow(workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func WorkflowsIdPut(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	workflowID, ok := parseWorkflowID(c)
	if !ok {
		return
	}

	var req requests.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid request body"})
		return
	}

	workflow, err := services.UpdateWorkflow(workflowID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": workflow})
}

func WorkflowsIdDelete(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	workflowID, ok := parseWorkflowID(c)
	if !ok {
		return
	}
	if err := services.DeleteWorkflow(workflowID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
