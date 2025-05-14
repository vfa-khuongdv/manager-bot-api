package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type IReminderScheduleHandler interface {
	CreateSchedule(ctx *gin.Context)
	GetSchedule(ctx *gin.Context)
	GetSchedulesByProject(ctx *gin.Context)
	UpdateSchedule(ctx *gin.Context)
	DeleteSchedule(ctx *gin.Context)
}

type ReminderScheduleHandler struct {
	service        services.IReminderScheduleService
	projectService services.IProjectService
	cronService    services.ICronService
}

// NewReminderScheduleHandler creates a new instance of ReminderScheduleHandler
// Parameters:
//   - service: The reminder schedule service to handle business logic
//   - projectService: The project service to validate project access
//   - cronService: The cron service to handle scheduled tasks
//
// Returns:
//   - *ReminderScheduleHandler: New handler instance
func NewReminderScheduleHandler(
	service services.IReminderScheduleService,
	projectService services.IProjectService,
	cronService services.ICronService,
) *ReminderScheduleHandler {
	return &ReminderScheduleHandler{
		service:        service,
		projectService: projectService,
		cronService:    cronService,
	}
}

// CreateSchedule handles the creation of a new reminder schedule
// It binds the JSON request body to a ReminderSchedule model and validates it
// before saving to the database
func (handler *ReminderScheduleHandler) CreateSchedule(ctx *gin.Context) {
	// Get secret key from header
	secretKey := ctx.GetHeader("X-Project-Secret-Key")
	if secretKey == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "Project secret key header is required"),
		)
		return
	}

	var input struct {
		ProjectID      uint   `json:"project_id" binding:"required"`
		CronExpression string `json:"cron_expression" binding:"required"`
		ChatworkRoomID string `json:"chatwork_room_id" binding:"required"`
		ChatworkToken  string `json:"chatwork_token" binding:"required"`
		Message        string `json:"message" binding:"required"`
		Active         bool   `json:"active" binding:"required"`
	}

	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, err.Error()),
		)
		return
	}

	// Validate project secret key
	isValid, err := handler.projectService.ValidateSecretKey(input.ProjectID, secretKey)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrResourceNotFound, "Project not found"),
		)
		return
	}

	if !isValid {
		utils.RespondWithError(
			ctx,
			http.StatusUnauthorized,
			errors.New(errors.ErrAuthUnauthorized, "Invalid project secret key"),
		)
		return
	}

	// Create a new reminder schedule model
	schedule := models.ReminderSchedule{
		ProjectID:      input.ProjectID,
		CronExpression: input.CronExpression,
		ChatworkRoomID: input.ChatworkRoomID,
		ChatworkToken:  input.ChatworkToken,
		Message:        input.Message,
		Active:         input.Active,
	}

	// Save to database
	if err := handler.service.Create(&schedule); err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Register with cron service if active
	if schedule.Active {
		handler.cronService.Register(&schedule)
	}

	utils.RespondWithOK(ctx, http.StatusCreated, schedule)
}

// GetSchedule retrieves a reminder schedule by ID
func (handler *ReminderScheduleHandler) GetSchedule(ctx *gin.Context) {
	// Get secret key from header
	secretKey := ctx.GetHeader("X-Project-Secret-Key")
	if secretKey == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "Project secret key header is required"),
		)
		return
	}

	id := ctx.Param("id")
	scheduleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	// Get the schedule
	schedule, err := handler.service.GetByID(uint(scheduleID))
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Validate project secret key
	isValid, err := handler.projectService.ValidateSecretKey(schedule.ProjectID, secretKey)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrResourceNotFound, "Project not found"),
		)
		return
	}

	if !isValid {
		utils.RespondWithError(
			ctx,
			http.StatusUnauthorized,
			errors.New(errors.ErrAuthUnauthorized, "Invalid project secret key"),
		)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{
		"schedule": utils.CensorSensitiveData(schedule, []string{"ChatworkToken"}),
	})
}

// GetSchedulesByProject retrieves all reminder schedules for a specific project
func (handler *ReminderScheduleHandler) GetSchedulesByProject(ctx *gin.Context) {
	// Get secret key from header
	secretKey := ctx.GetHeader("X-Project-Secret-Key")
	if secretKey == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "Project secret key header is required"),
		)
		return
	}

	projectID := ctx.Param("id")
	if projectID == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "project_id parameter is required"),
		)
		return
	}

	pID, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	// Validate project secret key before fetching schedules
	isValid, err := handler.projectService.ValidateSecretKey(uint(pID), secretKey)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrResourceNotFound, "Project not found"),
		)
		return
	}

	if !isValid {
		utils.RespondWithError(
			ctx,
			http.StatusUnauthorized,
			errors.New(errors.ErrAuthUnauthorized, "Invalid project secret key"),
		)
		return
	}

	schedules, err := handler.service.GetByProjectID(uint(pID))
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, schedules)
}

// UpdateSchedule handles updating an existing reminder schedule
func (handler *ReminderScheduleHandler) UpdateSchedule(ctx *gin.Context) {
	id := ctx.Param("id")
	scheduleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	// Get existing schedule
	existingSchedule, err := handler.service.GetByID(uint(scheduleID))
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Get secret key from header
	secretKey := ctx.GetHeader("X-Project-Secret-Key")
	if secretKey == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "Project secret key header is required"),
		)
		return
	}

	var input struct {
		ProjectID      uint   `json:"project_id"`
		CronExpression string `json:"cron_expression"`
		ChatworkRoomID string `json:"chatwork_room_id"`
		ChatworkToken  string `json:"chatwork_token"`
		Message        string `json:"message"`
		IsActive       *bool  `json:"is_active"`
	}

	// Bind and validate JSON request body
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, err.Error()),
		)
		return
	}

	// Validate project secret key
	isValid, err := handler.projectService.ValidateSecretKey(existingSchedule.ProjectID, secretKey)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrResourceNotFound, "Project not found"),
		)
		return
	}

	if !isValid {
		utils.RespondWithError(
			ctx,
			http.StatusUnauthorized,
			errors.New(errors.ErrAuthUnauthorized, "Invalid project secret key"),
		)
		return
	}

	// Update fields if provided
	if input.ProjectID != 0 {
		existingSchedule.ProjectID = input.ProjectID
	}
	if input.CronExpression != "" {
		existingSchedule.CronExpression = input.CronExpression
	}
	if input.ChatworkRoomID != "" {
		existingSchedule.ChatworkRoomID = input.ChatworkRoomID
	}
	if input.ChatworkToken != "" {
		existingSchedule.ChatworkToken = input.ChatworkToken
	}
	if input.Message != "" {
		existingSchedule.Message = input.Message
	}
	if input.IsActive != nil {
		existingSchedule.Active = *input.IsActive
	}

	// Save changes
	if err := handler.service.Update(existingSchedule); err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Update in cron service
	if existingSchedule.Active {
		fmt.Printf("Registering schedule with ID %d\n", existingSchedule.ID)
		handler.cronService.Register(existingSchedule)
	} else {
		handler.cronService.Remove(existingSchedule.ID)
	}

	utils.RespondWithOK(ctx, http.StatusOK, existingSchedule)
}

// DeleteSchedule handles deleting a reminder schedule
func (handler *ReminderScheduleHandler) DeleteSchedule(ctx *gin.Context) {
	// Get secret key from header
	secretKey := ctx.GetHeader("X-Project-Secret-Key")
	if secretKey == "" {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, "Project secret key header is required"),
		)
		return
	}

	id := ctx.Param("id")
	scheduleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	// Get the schedule to validate project access
	schedule, err := handler.service.GetByID(uint(scheduleID))
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Validate project secret key
	isValid, err := handler.projectService.ValidateSecretKey(schedule.ProjectID, secretKey)
	if err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			errors.New(errors.ErrResourceNotFound, "Project not found"),
		)
		return
	}

	if !isValid {
		utils.RespondWithError(
			ctx,
			http.StatusUnauthorized,
			errors.New(errors.ErrAuthUnauthorized, "Invalid project secret key"),
		)
		return
	}

	if err := handler.service.Delete(uint(scheduleID)); err != nil {
		utils.RespondWithError(
			ctx,
			http.StatusBadRequest,
			err,
		)
		return
	}

	// Remove from cron service
	handler.cronService.Remove(uint(scheduleID))

	var response struct {
		Message string `json:"message"`
	}
	response.Message = "Schedule deleted successfully"

	utils.RespondWithOK(ctx, http.StatusOK, response)
}
