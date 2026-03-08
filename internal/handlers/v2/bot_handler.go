package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type BotHandlerV2 struct {
	service services.IChatworkBotService
}

func NewBotHandlerV2(service services.IChatworkBotService) *BotHandlerV2 {
	return &BotHandlerV2{service: service}
}

// GET /api/v2/bots?page=1&limit=20
func (h *BotHandlerV2) GetAll(c *gin.Context) {
	paging := utils.GeneratePagingFromRequest(c)

	bots, total, err := h.service.GetAll(paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  bots,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

// POST /api/v2/bots
func (h *BotHandlerV2) Create(c *gin.Context) {
	var input struct {
		APIToken    string  `json:"apiToken" binding:"required"`
		Email       *string `json:"email"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "apiToken is required"))
		return
	}

	detail, err := h.service.Create(input.APIToken, input.Email, input.Description)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusCreated, detail)
}

// DELETE /api/v2/bots/:botId
func (h *BotHandlerV2) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "botId")
	if err != nil {
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
