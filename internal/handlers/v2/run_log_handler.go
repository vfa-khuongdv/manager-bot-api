package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// RunLogHandlerV2 handles V2 run log list endpoints
type RunLogHandlerV2 struct {
	service *services.ScheduleLogService
}

// NewRunLogHandlerV2 creates a new RunLogHandlerV2
func NewRunLogHandlerV2(service *services.ScheduleLogService) *RunLogHandlerV2 {
	return &RunLogHandlerV2{service: service}
}

// GetAll lists all run logs (admin only).
// GET /api/v2/run-logs?page=1&limit=20&status=success&projectId=1&scheduleId=1&from=2026-01-01&to=2026-03-31
func (h *RunLogHandlerV2) GetAll(c *gin.Context) {
	paging := utils.GeneratePagingFromRequest(c)
	filters := extractLogFilters(c)

	logs, total, err := h.service.ListAll(filters, paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// GetByProject lists run logs scoped to a specific project.
// GET /api/v2/projects/:projectId/run-logs
func (h *RunLogHandlerV2) GetByProject(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	paging := utils.GeneratePagingFromRequest(c)
	filters := extractLogFilters(c)
	filters["projectId"] = uint(projectID)

	logs, total, err := h.service.ListByProject(uint(projectID), filters, paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// GetBySchedule lists run logs for a specific schedule.
// GET /api/v2/projects/:projectId/schedules/:scheduleId/run-logs
func (h *RunLogHandlerV2) GetBySchedule(c *gin.Context) {
	_, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}
	scheduleID, err := parseIDParam(c, "scheduleId")
	if err != nil {
		return
	}

	paging := utils.GeneratePagingFromRequest(c)
	filters := extractLogFilters(c)

	logs, total, err := h.service.ListBySchedule(uint(scheduleID), filters, paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// extractLogFilters extracts common filter params from query string
func extractLogFilters(c *gin.Context) map[string]interface{} {
	filters := map[string]interface{}{}
	if s := c.Query("status"); s != "" {
		filters["status"] = s
	}
	if from := c.Query("from"); from != "" {
		filters["from"] = from
	}
	if to := c.Query("to"); to != "" {
		filters["to"] = to
	}
	return filters
}
