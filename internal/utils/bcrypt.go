package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword takes a plain text password string and returns a hashed version using bcrypt
// It uses the default cost factor for the hashing algorithm
// Returns the hashed password as a string and any error that occurred during hashing
func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return ""
	}

	return string(hashedPassword)
}

// CheckPasswordHash compares a plain text password with a hashed password
// Returns true if they match, false otherwise
func CheckPasswordHash(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
