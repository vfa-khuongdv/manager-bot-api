package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// RespondWithError sends a JSON error response with the given status code and error
// Parameters:
//   - ctx: Gin context for the request
//   - statusCode: HTTP status code to return
//   - err: Error to be included in response. If err is *errors.AppError, includes its code and message.
//     Otherwise includes internal error code and error message.
func RespondWithError(ctx *gin.Context, statusCode int, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		ctx.AbortWithStatusJSON(
			statusCode,
			gin.H{"code": appErr.Code, "message": appErr.Message},
		)
		return
	} else {
		ctx.AbortWithStatusJSON(
			statusCode,
			gin.H{"code": errors.ErrServerInternal, "message": err.Error()},
		)
		return
	}
}

// RespondWithOK sends a JSON response with the given status code and body
// Parameters:
//   - ctx: Gin context for the request
//   - statusCode: HTTP status code to return
//   - body: Data to be serialized as JSON response body
func RespondWithOK(ctx *gin.Context, statusCode int, body interface{}) {
	ctx.AbortWithStatusJSON(statusCode, body)
}
