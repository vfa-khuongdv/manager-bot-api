package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// APIKeyMiddleware validates API key from X-API-Key header or api_key query parameter
// The API key should be configured in API_KEY environment variable
func APIKeyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get expected API key from environment
		expectedAPIKey := utils.GetEnv("API_KEY", "")

		// If API_KEY is not configured, skip validation (for development)
		if expectedAPIKey == "" {
			ctx.Next()
			return
		}

		// Get API key from header or query parameter
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = ctx.Query("api_key")
		}

		// Validate API key
		if apiKey == "" {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "API key is required"))
			ctx.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Invalid API key"))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
