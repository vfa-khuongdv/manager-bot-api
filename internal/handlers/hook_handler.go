package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type HookHandler struct {
	cw      services.IChatworkService
	service services.IHookService
}

func NewHookHandler(cw services.IChatworkService, service services.IHookService) *HookHandler {
	return &HookHandler{
		cw:      cw,
		service: service,
	}
}

func (h *HookHandler) ChatworkHook(ctx *gin.Context) {
	var payload services.DiscordPayload

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		utils.RespondWithError(ctx, 400, err)
		return
	}

	logger.Infof("Received payload: %+v", payload)

	// Send the message to Chatwork
	err := h.service.ChatworkHook(payload)
	if err != nil {
		logger.Errorf("Failed to send message to Chatwork: %v", err)
		utils.RespondWithError(ctx, 500, err)
		return
	}

	utils.RespondWithOK(ctx, 200, gin.H{"status": "ok"})
}
