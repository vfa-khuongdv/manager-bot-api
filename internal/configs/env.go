package configs

import (
	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file.
// If no .env file is found, it will use system environment variables instead.
// Uses godotenv package to load the environment variables.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, loading environment variables from the system.")
	}
}
