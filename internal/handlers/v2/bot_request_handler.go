package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type BotRequestHandlerV2 struct {
	service services.IChatworkBotService
}

func NewBotRequestHandlerV2(service services.IChatworkBotService) *BotRequestHandlerV2 {
	return &BotRequestHandlerV2{service: service}
}

// GET /api/v2/bot-requests?status=pending
func (h *BotRequestHandlerV2) GetAll(c *gin.Context) {
	status := c.Query("status")

	items, err := h.service.GetBotRequests(status)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrServerInternal, err.Error()))
		return
	}
	if items == nil {
		items = []models.BotRequestItem{}
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  items,
		"total": len(items),
		"page":  1,
		"limit": len(items),
	})
}

// POST /api/v2/bot-requests/:requestId/accept
func (h *BotRequestHandlerV2) Accept(c *gin.Context) {
	compositeID := c.Param("requestId")

	if err := h.service.AcceptBotRequest(compositeID); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"success": true,
		"message": "Friend request accepted",
	})
}

// DELETE /api/v2/bot-requests/:requestId
func (h *BotRequestHandlerV2) Delete(c *gin.Context) {
	compositeID := c.Param("requestId")

	if err := h.service.DeleteBotRequest(compositeID); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
