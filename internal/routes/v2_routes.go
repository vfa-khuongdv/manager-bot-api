package routes

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/vfa-khuongdv/golang-cms/internal/handlers/v2"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

// SetupV2Routes registers all /api/v2 endpoints.
// Auth endpoints are public; all others require JWT Bearer token.
// Schedule endpoints additionally accept X-Project-Key header.
func SetupV2Routes(
	router *gin.Engine,
	projectService services.IProjectService,
	scheduleService services.IReminderScheduleService,
	logService *services.ScheduleLogService,
	cronService services.ICronService,
	chatworkService services.IChatworkService,
) {
	authHandler := v2.NewAuthHandler()
	projectHandler := v2.NewProjectHandlerV2(projectService, cronService)
	scheduleHandler := v2.NewScheduleHandlerV2(scheduleService, projectService, cronService, chatworkService)
	runLogHandler := v2.NewRunLogHandlerV2(logService)
	dashboardHandler := v2.NewDashboardHandlerV2(logService)

	apiV2 := router.Group("/api/v2")

	// ── Public: Auth ───────────────────────────────────────────────────────────
	apiV2.POST("/auth/login", authHandler.Login)
	apiV2.POST("/auth/logout", middlewares.JWTAuthMiddleware(), authHandler.Logout)

	// ── JWT-protected routes ───────────────────────────────────────────────────
	jwt := apiV2.Group("")
	jwt.Use(middlewares.JWTAuthMiddleware())
	{
		// Projects
		jwt.GET("/projects", projectHandler.GetAll)
		jwt.POST("/projects", projectHandler.Create)
		jwt.GET("/projects/:projectId", projectHandler.GetByID)
		jwt.PATCH("/projects/:projectId", projectHandler.Update)
		jwt.DELETE("/projects/:projectId", projectHandler.Delete)
		jwt.POST("/projects/:projectId/access", projectHandler.Access)

		// Dashboard
		jwt.GET("/dashboard/summary", dashboardHandler.GetSummary)

		// Run Logs (admin only — JWT required)
		jwt.GET("/run-logs", runLogHandler.GetAll)
		jwt.GET("/projects/:projectId/run-logs", runLogHandler.GetByProject)
		jwt.GET("/projects/:projectId/schedules/:scheduleId/run-logs", runLogHandler.GetBySchedule)
	}

	// ── Project-scoped routes (JWT or X-Project-Key) ───────────────────────────
	projectScoped := apiV2.Group("")
	projectScoped.Use(middlewares.ProjectScopeMiddleware())
	{
		projectScoped.GET("/projects/:projectId/schedules", scheduleHandler.GetByProject)
		projectScoped.POST("/projects/:projectId/schedules", scheduleHandler.Create)
		projectScoped.POST("/projects/:projectId/schedules/test", scheduleHandler.Test)
		projectScoped.GET("/projects/:projectId/schedules/:scheduleId", scheduleHandler.GetByID)
		projectScoped.PATCH("/projects/:projectId/schedules/:scheduleId", scheduleHandler.Update)
		projectScoped.PATCH("/projects/:projectId/schedules/:scheduleId/toggle", scheduleHandler.Toggle)
		projectScoped.DELETE("/projects/:projectId/schedules/:scheduleId", scheduleHandler.Delete)
	}
}
