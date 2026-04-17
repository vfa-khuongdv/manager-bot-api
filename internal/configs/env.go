package configs

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, loading environment variables from the system.")
	}
}

func GetFrontendURL() string {
	return os.Getenv("FRONTEND_URL")
}
