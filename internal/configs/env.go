package configs

import (
	"github.com/joho/godotenv"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

// LoadEnv loads environment variables from a .env file.
// If no .env file is found, it will use system environment variables instead.
// Uses godotenv package to load the environment variables.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, loading environment variables from the system.")
	}
}
