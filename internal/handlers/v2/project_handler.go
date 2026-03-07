package v2

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

// ProjectHandlerV2 handles V2 project endpoints
type ProjectHandlerV2 struct {
	service     services.IProjectService
	cronService services.ICronService
}

// NewProjectHandlerV2 creates a new ProjectHandlerV2
func NewProjectHandlerV2(service services.IProjectService, cronService services.ICronService) *ProjectHandlerV2 {
	return &ProjectHandlerV2{
		service:     service,
		cronService: cronService,
	}
}

// GetAll lists all projects with pagination and optional status filter.
// GET /api/v2/projects?page=1&limit=20&status=active
func (h *ProjectHandlerV2) GetAll(c *gin.Context) {
	paging := utils.GeneratePagingFromRequest(c)
	statusFilter := c.Query("status") // "active" | "inactive" | ""

	projects, total, err := h.service.GetAllV2(statusFilter, paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	projectResponses := make([]gin.H, 0, len(projects))
	for i := range projects {
		projectResponses = append(projectResponses, buildProjectResponse(&projects[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  projectResponses,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// Create creates a new project. secretKey is auto-generated server-side.
// POST /api/v2/projects
func (h *ProjectHandlerV2) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "name is required"))
		return
	}

	status := input.Status
	if status == "" {
		status = "active"
	}

	// Auto-generate secret key: sk_proj_<12 random chars>
	secretKey := fmt.Sprintf("sk_proj_%s", utils.GenerateRandomString(12))

	project := &models.Project{
		Name:        input.Name,
		Description: input.Description,
		Status:      status,
		SecretKey:   secretKey,
	}

	created, err := h.service.Create(project)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	// Return secretKey in create response (only time it is visible)
	utils.RespondWithOK(c, http.StatusCreated, gin.H{
		"id":             created.ID,
		"name":           created.Name,
		"description":    created.Description,
		"status":         created.Status,
		"secretKey":      created.SecretKey,
		"createdAt":      created.CreatedAt.Format("2006-01-02"),
		"schedulesCount": 0,
	})
}

// GetByID returns a single project.
// GET /api/v2/projects/:projectId
func (h *ProjectHandlerV2) GetByID(c *gin.Context) {
	id, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Project not found"))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, buildProjectResponse(project))
}

// Update partially updates a project. secretKey cannot be changed.
// PATCH /api/v2/projects/:projectId
func (h *ProjectHandlerV2) Update(c *gin.Context) {
	id, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Project not found"))
		return
	}

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	if input.Name != nil {
		project.Name = *input.Name
	}
	if input.Description != nil {
		project.Description = *input.Description
	}
	if input.Status != nil {
		project.Status = *input.Status
	}

	updated, err := h.service.Update(project)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	h.cronService.SyncAll()

	utils.RespondWithOK(c, http.StatusOK, buildProjectResponse(updated))
}

// Delete deletes a project and all its schedules.
// DELETE /api/v2/projects/:projectId
func (h *ProjectHandlerV2) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	_, err = h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Project not found"))
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	h.cronService.SyncAll()

	c.Status(http.StatusNoContent)
}

// Access validates a project's secret key.
// POST /api/v2/projects/:projectId/access
func (h *ProjectHandlerV2) Access(c *gin.Context) {
	id, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	var input struct {
		SecretKey string `json:"secretKey" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "secretKey is required"))
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Project not found"))
		return
	}

	valid, err := h.service.ValidateSecretKey(uint(id), input.SecretKey)
	if err != nil || !valid {
		utils.RespondWithError(c, http.StatusForbidden, errors.New(errors.ErrAuthForbidden, "Invalid secret key"))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"projectId": project.ID,
		"name":      project.Name,
		"granted":   true,
	})
}

// ---- helpers ----

func parseIDParam(c *gin.Context, param string) (int, error) {
	raw := c.Param(param)
	id, err := strconv.Atoi(raw)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParse, "Invalid ID parameter"))
		return 0, err
	}
	return id, nil
}

func buildProjectResponse(p *models.Project) gin.H {
	schedulesCount := p.SchedulesCount
	if schedulesCount == 0 {
		schedulesCount = len(p.ReminderSchedules)
	}
	return gin.H{
		"id":             p.ID,
		"name":           p.Name,
		"description":    p.Description,
		"status":         p.Status,
		"createdAt":      p.CreatedAt.Format("2006-01-02"),
		"schedulesCount": schedulesCount,
	}
}
