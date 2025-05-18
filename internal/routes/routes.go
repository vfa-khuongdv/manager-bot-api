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

	// Project
	projectRepo := repositories.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepo)
	projectHandler := handlers.NewProjectHandler(projectService, cronService)

	// Reminder
	reminderScheduleRepo := repositories.NewReminderScheduleRepository(db)
	reminderScheduleService := services.NewReminderScheduleService(reminderScheduleRepo)

	reminderScheduleHandler := handlers.NewReminderScheduleHandler(
		reminderScheduleService,
		projectService,
		cronService,
	)

	// Hook handler
	chatworkService := services.NewChatworkService()
	hookService := services.NewHookService(chatworkService)

	hookHandler := handlers.NewHookHandler(chatworkService, hookService)

	// Add middleware for CORS and logging
	router.Use(
		middlewares.CORSMiddleware(),
		middlewares.LogMiddleware(),
		gin.Recovery(),
		middlewares.EmptyBodyMiddleware(),
	)

	router.GET("/healthz", handlers.HealthCheck)

	// Setup API routes
	api := router.Group("/api/v1")
	{
		// Project routes

		api.GET("/projects", projectHandler.GetAll)
		api.POST("/projects", projectHandler.Create)
		api.GET("/projects/:id", projectHandler.GetByID)
		api.PATCH("/projects/:id", projectHandler.Update)
		api.DELETE("/projects/:id", projectHandler.Delete)
		api.POST("/projects/verify-access", projectHandler.VerifyAccess)

		// Reminder schedule routes
		api.POST("/reminder-schedules", reminderScheduleHandler.CreateSchedule)
		api.GET("/reminder-schedules/:id", reminderScheduleHandler.GetSchedule)                    // This :id is for the reminder schedule itself
		api.GET("/projects/:id/reminder-schedules", reminderScheduleHandler.GetSchedulesByProject) // Changed :project_id to :id
		api.PATCH("/reminder-schedules/:id", reminderScheduleHandler.UpdateSchedule)
		api.DELETE("/reminder-schedules/:id", reminderScheduleHandler.DeleteSchedule)

		// Test hooks
		api.POST("/hooks/chatwork", hookHandler.ChatworkHook)
	}

	return router
}
