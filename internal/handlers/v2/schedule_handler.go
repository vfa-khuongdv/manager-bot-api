package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// ScheduleHandlerV2 handles V2 schedule endpoints
type ScheduleHandlerV2 struct {
	service         services.IReminderScheduleService
	projectService  services.IProjectService
	cronService     services.ICronService
	chatworkService services.IChatworkService
}

// NewScheduleHandlerV2 creates a new ScheduleHandlerV2
func NewScheduleHandlerV2(
	service services.IReminderScheduleService,
	projectService services.IProjectService,
	cronService services.ICronService,
	chatworkService services.IChatworkService,
) *ScheduleHandlerV2 {
	return &ScheduleHandlerV2{
		service:         service,
		projectService:  projectService,
		cronService:     cronService,
		chatworkService: chatworkService,
	}
}

// GetByProject lists all schedules for a project.
// GET /api/v2/projects/:projectId/schedules?page=1&limit=20&status=active
func (h *ScheduleHandlerV2) GetByProject(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	paging := utils.GeneratePagingFromRequest(c)
	statusFilter := c.Query("status")

	schedules, total, err := h.service.GetByProjectIDPaged(uint(projectID), statusFilter, paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	data := make([]gin.H, 0, len(schedules))
	for _, s := range schedules {
		data = append(data, buildScheduleResponse(&s))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// Create creates a new schedule.
// POST /api/v2/projects/:projectId/schedules
func (h *ScheduleHandlerV2) Create(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	var input struct {
		Name    string `json:"name" binding:"required"`
		RoomID  string `json:"roomId" binding:"required"`
		APIKey  string `json:"apiKey" binding:"required"`
		Cron    string `json:"cron" binding:"required"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	active := input.Status != "paused"

	schedule := models.ReminderSchedule{
		ProjectID:      uint(projectID),
		Name:           input.Name,
		CronExpression: input.Cron,
		ChatworkRoomID: input.RoomID,
		ChatworkToken:  input.APIKey,
		Message:        input.Message,
		Active:         active,
	}

	if err := h.service.Create(&schedule); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseInsert, err.Error()))
		return
	}

	if schedule.Active {
		h.cronService.Register(&schedule)
	}

	utils.RespondWithOK(c, http.StatusCreated, buildScheduleResponse(&schedule))
}

// GetByID gets a single schedule.
// GET /api/v2/projects/:projectId/schedules/:scheduleId
func (h *ScheduleHandlerV2) GetByID(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}
	scheduleID, err := parseIDParam(c, "scheduleId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	schedule, err := h.service.GetByID(uint(scheduleID))
	if err != nil || schedule == nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found"))
		return
	}

	if schedule.ProjectID != uint(projectID) {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found in this project"))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, buildScheduleResponse(schedule))
}

// Update partially updates a schedule. Omitting apiKey leaves existing key unchanged.
// PATCH /api/v2/projects/:projectId/schedules/:scheduleId
func (h *ScheduleHandlerV2) Update(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}
	scheduleID, err := parseIDParam(c, "scheduleId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	schedule, err := h.service.GetByID(uint(scheduleID))
	if err != nil || schedule == nil || schedule.ProjectID != uint(projectID) {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found"))
		return
	}

	var input struct {
		Name    *string `json:"name"`
		RoomID  *string `json:"roomId"`
		APIKey  *string `json:"apiKey"`
		Cron    *string `json:"cron"`
		Message *string `json:"message"`
		Status  *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	if input.Name != nil {
		schedule.Name = *input.Name
	}
	if input.RoomID != nil {
		schedule.ChatworkRoomID = *input.RoomID
	}
	if input.APIKey != nil {
		schedule.ChatworkToken = *input.APIKey
	}
	if input.Cron != nil {
		schedule.CronExpression = *input.Cron
	}
	if input.Message != nil {
		schedule.Message = *input.Message
	}
	if input.Status != nil {
		schedule.Active = *input.Status != "paused"
	}

	if err := h.service.Update(schedule); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseUpdate, err.Error()))
		return
	}

	if schedule.Active {
		h.cronService.Register(schedule)
	} else {
		h.cronService.Remove(schedule.ID)
	}

	utils.RespondWithOK(c, http.StatusOK, buildScheduleResponse(schedule))
}

// Toggle toggles schedule status between active and paused.
// PATCH /api/v2/projects/:projectId/schedules/:scheduleId/toggle
func (h *ScheduleHandlerV2) Toggle(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}
	scheduleID, err := parseIDParam(c, "scheduleId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	schedule, err := h.service.GetByID(uint(scheduleID))
	if err != nil || schedule == nil || schedule.ProjectID != uint(projectID) {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found"))
		return
	}

	newActive := !schedule.Active
	if err := h.service.ToggleActiveStatus(uint(scheduleID), newActive); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseUpdate, err.Error()))
		return
	}

	if newActive {
		schedule.Active = true
		h.cronService.Register(schedule)
	} else {
		h.cronService.Remove(schedule.ID)
	}

	newStatus := "active"
	if !newActive {
		newStatus = "paused"
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"id":     schedule.ID,
		"status": newStatus,
	})
}

// Delete deletes a schedule.
// DELETE /api/v2/projects/:projectId/schedules/:scheduleId
func (h *ScheduleHandlerV2) Delete(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}
	scheduleID, err := parseIDParam(c, "scheduleId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	schedule, err := h.service.GetByID(uint(scheduleID))
	if err != nil || schedule == nil || schedule.ProjectID != uint(projectID) {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found"))
		return
	}

	if err := h.service.Delete(uint(scheduleID)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseDelete, err.Error()))
		return
	}

	h.cronService.Remove(uint(scheduleID))
	c.Status(http.StatusNoContent)
}

// Test sends a test message to Chatwork using the provided parameters.
// POST /api/v2/projects/:projectId/schedules/test
func (h *ScheduleHandlerV2) Test(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	var input struct {
		RoomID     string `json:"roomId" binding:"required"`
		APIKey     string `json:"apiKey" binding:"required"`
		Message    string `json:"message" binding:"required"`
		ScheduleID *uint  `json:"scheduleId"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	apiKey := input.APIKey
	// If the frontend sends the masked API key, we must look it up from the database using ScheduleID
	if apiKey == "cwk_***hidden***" {
		if input.ScheduleID == nil {
			utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "scheduleId is required when apiKey is masked"))
			return
		}

		schedule, err := h.service.GetByID(*input.ScheduleID)
		if err != nil || schedule == nil || schedule.ProjectID != uint(projectID) {
			utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Schedule not found to resolve API key"))
			return
		}
		apiKey = schedule.ChatworkToken
	}

	err = h.chatworkService.SendMessage(apiKey, input.RoomID, input.Message)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadGateway, errors.New(errors.ErrServerInternal, "Failed to send message: "+err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"success": true,
		"message": "Test message sent successfully",
	})
}

// ---- helpers ----

// checkProjectAccess validates project-scoped access using either JWT (admin) or X-Project-Key header.
func (h *ScheduleHandlerV2) checkProjectAccess(c *gin.Context, projectID uint) error {
	// If JWT-authenticated (admin), skip secret key check
	if authMode, _ := c.Get("authMode"); authMode == "jwt" {
		return nil
	}

	// Otherwise validate X-Project-Key
	projectKey, _ := c.Get("projectKey")
	keyStr, ok := projectKey.(string)
	if !ok || keyStr == "" {
		utils.RespondWithError(c, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Project key required"))
		return errors.New(errors.ErrAuthUnauthorized, "missing project key")
	}

	valid, err := h.projectService.ValidateSecretKey(projectID, keyStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Project not found"))
		return err
	}
	if !valid {
		utils.RespondWithError(c, http.StatusForbidden, errors.New(errors.ErrAuthForbidden, "Invalid project secret key"))
		return errors.New(errors.ErrAuthForbidden, "invalid project key")
	}
	return nil
}

// buildScheduleResponse converts a ReminderSchedule to V2 response format.
// apiKey is always masked as "cwk_***hidden***".
func buildScheduleResponse(s *models.ReminderSchedule) gin.H {
	status := "active"
	if !s.Active {
		status = "paused"
	}

	projectName := ""
	if s.Project.Name != "" {
		projectName = s.Project.Name
	}

	lastRun := interface{}(nil)
	lastStatus := interface{}(nil)
	if !s.UpdatedAt.IsZero() && s.UpdatedAt != s.CreatedAt {
		lastRun = s.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	_ = lastStatus // populated by run logs in real impl

	return gin.H{
		"id":          s.ID,
		"projectId":   s.ProjectID,
		"name":        s.Name,
		"projectName": projectName,
		"roomId":      s.ChatworkRoomID,
		"apiKey":      "cwk_***hidden***",
		"cron":        s.CronExpression,
		"message":     s.Message,
		"status":      status,
		"lastRun":     lastRun,
		"lastStatus":  lastStatus,
		"createdAt":   s.CreatedAt.Format("2006-01-02"),
	}
}
