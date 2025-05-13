package seeders

import (
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

// Run executes all seed functions to populate the database with initial data
// It takes a GORM database connection as input and panics if any seeding operation fails
func Run(db *gorm.DB) {

	logger.Info("Seeding database...")

}
