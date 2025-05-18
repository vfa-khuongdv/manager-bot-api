package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type HookHandler struct {
	cw services.IChatworkService
}

func NewHookHandler(cw services.IChatworkService) *HookHandler {
	return &HookHandler{
		cw: cw,
	}
}

func (h *HookHandler) ChatworkHook(ctx *gin.Context) {
	// print the request body
	// print the query params
	logger.Infof("Request Body: %s", ctx.Request.Body)
	logger.Infof("Query Params: %v", ctx.Request.URL.Query())
	utils.RespondWithOK(ctx, 200, gin.H{"status": "ok"})
}
