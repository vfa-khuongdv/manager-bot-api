package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// Middleware to reject requests with empty JSON body (except for specified routes)
func EmptyBodyMiddleware() gin.HandlerFunc {
	skipRouteSuffixes := []string{"/scan", "/toggle", "/test", "/run"}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			shouldSkip := false
			path := c.Request.URL.Path
			for _, suffix := range skipRouteSuffixes {
				if strings.HasSuffix(path, suffix) {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				c.Next()
				return
			}

			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil || len(bytes.TrimSpace(bodyBytes)) == 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    errors.ErrInvalidData,
					"message": "Request body cannot be empty",
				})
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		c.Next()
	}
}
