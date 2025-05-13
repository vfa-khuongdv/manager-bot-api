package utils

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// ValidateBirthday checks if the birthday is in a valid format and not a future date.
func ValidateBirthday(fl validator.FieldLevel) bool {
	birthdayStr := fl.Field().String()
	layout := "2006-01-02" // Format: YYYY-MM-DD

	// Parse the birthday to check the format
	parsedDate, err := time.Parse(layout, birthdayStr)
	if err != nil {
		return false // Invalid date format
	}

	// Check if the birthday is in the future
	if parsedDate.After(time.Now()) {
		return false // Invalid: birthday can't be in the future
	}

	return true // Valid birthday
}
