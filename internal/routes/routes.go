package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cronService *services.CronService) *gin.Engine {
	// Set Gin mode from environment variable
	ginMode := utils.GetEnv("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	// Initialize the default Gin router
	router := gin.Default()

	// Repository
	projectRepo := repositories.NewProjectRepository(db)
	reminderScheduleRepo := repositories.NewReminderScheduleRepository(db)
	scheduleLogRepo := repositories.NewScheduleLogRepository(db)
	chatworkBotRepo := repositories.NewChatworkBotRepository(db)

	// Services
	projectService := services.NewProjectService(projectRepo)
	reminderScheduleService := services.NewReminderScheduleService(reminderScheduleRepo)
	chatworkService := services.NewChatworkService()
	hookService := services.NewHookService(chatworkService)
	scheduleLogService := services.NewScheduleLogService(scheduleLogRepo)
	botService := services.NewChatworkBotService(chatworkBotRepo, reminderScheduleRepo)

	// Handlers
	hookHandler := handlers.NewHookHandler(chatworkService, hookService)

	// Add middleware for CORS and logging
	router.Use(
		middlewares.CORSMiddleware(),
		middlewares.LogMiddleware(),
		gin.Recovery(),
		middlewares.EmptyBodyMiddleware(),
	)

	// Health check routes
	router.GET("/healthz", handlers.HealthCheck)
	router.GET("/readyz", handlers.Test)

	// Setup API routes with API key authentication
	api := router.Group("/api/v1")

	// Hook routes
	api.POST("/hooks/chatwork", hookHandler.ChatworkHook)
	api.POST("/hooks/slack", hookHandler.SlackHook)

	// Setup V2 routes
	SetupV2Routes(router, projectService, reminderScheduleService, scheduleLogService, cronService, chatworkService, botService)

	return router
}
