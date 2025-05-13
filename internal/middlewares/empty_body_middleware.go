package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// Middleware to reject requests with empty JSON body
func EmptyBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil || len(bytes.TrimSpace(bodyBytes)) == 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    errors.ErrInvalidData,
					"message": "Request body cannot be empty",
				})
				return
			}
			// Replace the body so the handler can read it again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		c.Next()
	}
}
