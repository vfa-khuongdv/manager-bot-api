package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func HealthCheck(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "healthy"})
}

func Test(ctx *gin.Context) {
	testEnv := utils.GetEnv("TEST_ENV", "not set")
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"data": testEnv})
}

func GetHealth(ctx *gin.Context) {
	health := services.GetOverallHealth()
	utils.RespondWithOK(ctx, http.StatusOK, health)
}

func GetChatworkHealth(ctx *gin.Context) {
	chatworkService := services.NewHealthChatworkService()
	health, err := chatworkService.CheckHealth()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":   "down",
			"message": err.Error(),
		})
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, health)
}

func GetServerHealth(ctx *gin.Context) {
	health := services.GetServerHealth()
	utils.RespondWithOK(ctx, http.StatusOK, health)
}

func GetDatabaseHealth(ctx *gin.Context) {
	health, err := services.GetDatabaseHealth()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":   "disconnected",
			"message": err.Error(),
		})
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, health)
}
