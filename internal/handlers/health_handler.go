package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func HealthCheck(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "healthy"})
}

func Test(ctx *gin.Context) {
	testEnv := utils.GetEnv("TEST_ENV", "not set")
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"data": testEnv})
}
