package v2

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// AuthHandler handles V2 authentication endpoints
type AuthHandler struct{}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login validates the admin passcode and returns a JWT token.
// POST /api/v2/auth/login
// Body: { "passcode": "..." }
// Response 200: { "token": "...", "expiresAt": "2026-03-07T..." }
// Response 401: Invalid passcode
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Passcode string `json:"passcode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "passcode is required"))
		return
	}

	expectedPasscode := utils.GetEnv("ADMIN_PASSCODE", "botadmin2026")
	if input.Passcode != expectedPasscode {
		utils.RespondWithError(c, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Invalid passcode"))
		return
	}

	// Generate JWT — use ID=0 for the admin session (passcode-based, no user record)
	result, err := configs.GenerateToken(0)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrServerInternal, "Failed to generate token"))
		return
	}

	expiresAt := time.Unix(result.ExpiresAt, 0).UTC().Format(time.RFC3339)
	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"token":     result.Token,
		"expiresAt": expiresAt,
	})
}

// Logout invalidates the current session (client-side; stateless JWT).
// POST /api/v2/auth/logout
// Response 204: No content
func (h *AuthHandler) Logout(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
