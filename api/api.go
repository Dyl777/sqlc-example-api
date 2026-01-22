package api

import (
	"net/http"

	"github.com/Iknite-Space/sqlc-example-api/db/repo"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	querier repo.Querier
}

func NewMessageHandler(querier repo.Querier) *MessageHandler {
	return &MessageHandler{
		querier: querier,
	}
}

func (h *MessageHandler) WireHttpHandler() http.Handler {
	r := gin.Default()
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.String(http.StatusInternalServerError, "Internal Server Error: panic")
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// Original message endpoints
	r.POST("/message", h.handleCreateMessage)
	r.GET("/message/:id", h.handleGetMessage)
	r.GET("/thread/:id/messages", h.handleGetThreadMessages)

	// Dashboard endpoints
	dashboardHandler := NewDashboardHandler(h.querier)
	dashboard := r.Group("/dashboard")
	{
		dashboard.GET("/summary", dashboardHandler.handleGetDashboardSummary)

		// Docker containers
		dashboard.POST("/containers", dashboardHandler.handleCreateDockerContainer)
		dashboard.GET("/containers", dashboardHandler.handleListDockerContainers)
		dashboard.GET("/containers/:id", dashboardHandler.handleGetDockerContainer)
		dashboard.PUT("/containers/:id", dashboardHandler.handleUpdateDockerContainer)

		// Git repositories
		dashboard.POST("/repos", dashboardHandler.handleCreateGitRepo)
		dashboard.GET("/repos", dashboardHandler.handleListGitRepos)
	}

	// System Data endpoints
	systemDataHandler := NewSystemDataHandler(h.querier)
	system := r.Group("/system")
	{
		// Cache data
		system.POST("/cache", systemDataHandler.handleCreateCacheData)
		system.GET("/cache", systemDataHandler.handleListCacheData)
		system.GET("/cache/:id", systemDataHandler.handleGetCacheData)

		// Log entries
		system.POST("/logs", systemDataHandler.handleCreateLogEntry)
		system.GET("/logs", systemDataHandler.handleListLogEntries)
		system.GET("/logs/:id", systemDataHandler.handleGetLogEntry)

		// Secrets
		system.POST("/secrets", systemDataHandler.handleCreateSecret)
		system.GET("/secrets", systemDataHandler.handleListSecrets)
		system.GET("/secrets/:id", systemDataHandler.handleGetSecret)

		// Registry data (Windows)
		system.POST("/registry", systemDataHandler.handleCreateRegistryData)
		system.GET("/registry", systemDataHandler.handleListRegistryData)
		system.GET("/registry/:id", systemDataHandler.handleGetRegistryData)

		// Plist data (macOS)
		system.POST("/plist", systemDataHandler.handleCreatePlistData)
		system.GET("/plist", systemDataHandler.handleListPlistData)
		system.GET("/plist/:id", systemDataHandler.handleGetPlistData)
	}

	// Schema management endpoints
	schemaHandler := NewSchemaHandler(h.querier)
	schema := r.Group("/schema")
	{
		// Schema definition management
		schema.GET("/:table", schemaHandler.handleGetTableSchema)
		schema.POST("/:table", schemaHandler.handleCreateTableSchema)
		schema.GET("/:table/versions", schemaHandler.handleListTableSchemas)

		// Dynamic field management
		schema.POST("/:table/:id/fields", schemaHandler.handleAddFieldToRecord)
		schema.DELETE("/:table/:id/fields/:field", schemaHandler.handleRemoveFieldFromRecord)

		// Search by field values
		schema.POST("/:table/search", schemaHandler.handleSearchByField)
	}

	// Workflow endpoints
	workflowHandler := NewWorkflowHandler(h.querier)
	workflow := r.Group("/workflow")
	{
		workflow.POST("/", workflowHandler.handleCreateWorkflow)
		workflow.GET("/", workflowHandler.handleListWorkflows)
		workflow.GET("/:id", workflowHandler.handleGetWorkflow)
		workflow.POST("/import", workflowHandler.handleImportWorkflowData)

		// Editor configs
		workflow.POST("/config", workflowHandler.handleCreateEditorConfig)
		workflow.GET("/config", workflowHandler.handleListEditorConfigs)
		workflow.GET("/config/:id", workflowHandler.handleGetEditorConfig)
	}

	return r
}

func (h *MessageHandler) handleCreateMessage(c *gin.Context) {
	var req repo.CreateMessageParams
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.querier.CreateMessage(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *MessageHandler) handleGetMessage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	message, err := h.querier.GetMessageByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message)
}

func (h *MessageHandler) handleGetThreadMessages(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	messages, err := h.querier.GetMessagesByThread(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"thread":   id,
		"topic":    "example",
		"messages": messages,
	})
}
