package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type ScheduleLogHandler struct {
	service *services.ScheduleLogService
}

func NewScheduleLogHandler(service *services.ScheduleLogService) *ScheduleLogHandler {
	return &ScheduleLogHandler{service: service}
}

func (h *ScheduleLogHandler) GetDashboard(c *gin.Context) {
	data, err := h.service.GetDashboardData()
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	utils.RespondWithOK(c, http.StatusOK, data)
}
