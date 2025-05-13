package configs

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type CustomClaims struct {
	ID                   uint `json:"id"` // Custom field
	jwt.RegisteredClaims      // // Embed standard claims
}

var jwtKey = []byte(utils.GetEnv("JWT_KEY", "replace_your_key"))

type JwtResult struct {
	Token     string
	ExpiresAt int64
}

// GenerateToken creates a new JWT token for the given email
// Parameters:
//   - id: the userId to be included in the token claims
//
// Returns:
//   - *JwtResult: contains the signed token string and expiration timestamp
//   - error: any error that occurred during token generation
func GenerateToken(id uint) (*JwtResult, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour))
	claims := CustomClaims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString(jwtKey)

	if err != nil {
		return nil, err
	}

	return &JwtResult{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ValidateToken validates a JWT token string and extracts the claims
// Parameters:
//   - tokenString: the JWT token string to validate
//
// Returns:
//   - *CustomClaims: the extracted claims if token is valid
//   - error: any error that occurred during validation
func ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err

}
