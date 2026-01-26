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

	// Services
	projectService := services.NewProjectService(projectRepo)
	reminderScheduleService := services.NewReminderScheduleService(reminderScheduleRepo)
	chatworkService := services.NewChatworkService()
	hookService := services.NewHookService(chatworkService)
	scheduleLogService := services.NewScheduleLogService(scheduleLogRepo)

	// Handlers
	projectHandler := handlers.NewProjectHandler(projectService, cronService)
	hookHandler := handlers.NewHookHandler(chatworkService, hookService)
	reminderScheduleHandler := handlers.NewReminderScheduleHandler(
		reminderScheduleService,
		projectService,
		cronService,
	)
	scheduleLogHandler := handlers.NewScheduleLogHandler(scheduleLogService)

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
	api.Use(middlewares.APIKeyMiddleware())
	{
		// Project routes
		api.GET("/projects", projectHandler.GetAll)
		api.POST("/projects", projectHandler.Create)
		api.GET("/projects/:id", projectHandler.GetByID)
		api.PATCH("/projects/:id", projectHandler.Update)
		api.DELETE("/projects/:id", projectHandler.Delete)
		api.POST("/projects/verify-access", projectHandler.VerifyAccess)

		// Reminder routes
		api.POST("/reminder-schedules", reminderScheduleHandler.CreateSchedule)
		api.GET("/reminder-schedules/:id", reminderScheduleHandler.GetSchedule)                    // This :id is for the reminder schedule itself
		api.GET("/projects/:id/reminder-schedules", reminderScheduleHandler.GetSchedulesByProject) // Changed :project_id to :id
		api.PATCH("/reminder-schedules/:id", reminderScheduleHandler.UpdateSchedule)
		api.DELETE("/reminder-schedules/:id", reminderScheduleHandler.DeleteSchedule)

		// Hook routes
		api.POST("/hooks/chatwork", hookHandler.ChatworkHook)
		api.POST("/hooks/slack", hookHandler.SlackHook)

		// Dashboard routes
		api.GET("/dashboard", scheduleLogHandler.GetDashboard)
	}

	return router
}
