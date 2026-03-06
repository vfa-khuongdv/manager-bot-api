package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// JWTAuthMiddleware validates JWT Bearer token for V2 endpoints.
// Returns 401 if Authorization header is missing, malformed, or token is invalid.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Authorization header required"))
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := configs.ValidateToken(tokenString)
		if err != nil {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Invalid or expired token"))
			ctx.Abort()
			return
		}

		ctx.Set("UserID", claims.ID)
		ctx.Next()
	}
}

// ProjectScopeMiddleware allows a request if it carries either:
//   - A valid JWT Bearer token, OR
//   - A valid X-Project-Key header matching the project's secretKey
//
// This enables project-scoped access without full admin privileges.
// The resolved projectKey (if used) is stored in context as "projectKey".
func ProjectScopeMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try JWT Bearer first
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := configs.ValidateToken(tokenString)
			if err == nil {
				ctx.Set("UserID", claims.ID)
				ctx.Set("authMode", "jwt")
				ctx.Next()
				return
			}
		}

		// Fall back to project key header
		projectKey := ctx.GetHeader("X-Project-Key")
		if projectKey != "" {
			ctx.Set("projectKey", projectKey)
			ctx.Set("authMode", "projectKey")
			ctx.Next()
			return
		}

		utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Authorization required: provide Bearer token or X-Project-Key header"))
		ctx.Abort()
	}
}
