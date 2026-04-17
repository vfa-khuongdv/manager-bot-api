package v2

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// DashboardHandlerV2 handles V2 dashboard endpoints
type DashboardHandlerV2 struct {
	scheduleLogService *services.ScheduleLogService
	cveConfigService   services.ICveConfigService
}

// NewDashboardHandlerV2 creates a new DashboardHandlerV2
func NewDashboardHandlerV2(scheduleLogService *services.ScheduleLogService, cveConfigService services.ICveConfigService) *DashboardHandlerV2 {
	return &DashboardHandlerV2{
		scheduleLogService: scheduleLogService,
		cveConfigService:   cveConfigService,
	}
}

// GetSummary returns aggregated dashboard stats.
// GET /api/v2/dashboard/summary
// Response: { activeProjects, inactiveProjects, totalSchedules, activeSchedules, successRuns, failedRuns, successRate, totalCveConfigs, activeCveMonitoring, totalVulnerabilities, secureConfigs, criticalVulns, highVulns, moderateVulns, lowVulns }
func (h *DashboardHandlerV2) GetSummary(c *gin.Context) {
	summary, err := h.scheduleLogService.GetV2Summary()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}
	utils.RespondWithOK(c, http.StatusOK, summary)
}

// GetCveRecentScans returns recent CVE scan results.
// GET /api/v2/dashboard/cve-recent-scans?limit=10
func (h *DashboardHandlerV2) GetCveRecentScans(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	results, total, err := h.cveConfigService.GetRecentScans(limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  results,
		"total": total,
	})
}
