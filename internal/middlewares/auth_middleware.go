package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// AuthMiddleware is a Gin middleware function that handles JWT authentication
// It validates the Authorization header and extracts the JWT token
// The middleware checks if:
// - Authorization header exists and has "Bearer " prefix
// - Token is valid and can be parsed
// If validation succeeds, it sets the user ID from token claims in context
// If validation fails, it returns 401 Unauthorized
func AuthMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Authorization header required"))
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := configs.ValidateToken(tokenString)
		if err != nil {
			utils.RespondWithError(ctx, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Unauthorized"))
		}

		ctx.Set("UserID", claims.ID)
		ctx.Next()
	}
}
