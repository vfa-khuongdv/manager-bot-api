package utils

import (
	"math/rand"
	"time"
)

// GenerateRandomString generates a random string of specified length using alphanumeric characters
// Parameters:
//   - n: length of the random string to generate
//
// Returns:
//   - string: randomly generated alphanumeric string of length n
func GenerateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return generateStringWithCharset(n, charset)
}

// generateStringWithCharset generates a random string of specified length using provided character set
// Parameters:
//   - n: length of the random string to generate
//   - charset: string containing the characters to use for random generation
//
// Returns:
//   - string: randomly generated string of length n using characters from charset
func generateStringWithCharset(n int, charset string) string {
	// Create a new random generator with a random seed
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, n)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}
