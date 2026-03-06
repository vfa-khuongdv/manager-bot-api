package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// DashboardHandlerV2 handles V2 dashboard endpoints
type DashboardHandlerV2 struct {
	service *services.ScheduleLogService
}

// NewDashboardHandlerV2 creates a new DashboardHandlerV2
func NewDashboardHandlerV2(service *services.ScheduleLogService) *DashboardHandlerV2 {
	return &DashboardHandlerV2{service: service}
}

// GetSummary returns aggregated dashboard stats.
// GET /api/v2/dashboard/summary
// Response: { activeProjects, inactiveProjects, totalSchedules, activeSchedules, successRuns, failedRuns, successRate }
func (h *DashboardHandlerV2) GetSummary(c *gin.Context) {
	summary, err := h.service.GetV2Summary()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}
	utils.RespondWithOK(c, http.StatusOK, summary)
}
