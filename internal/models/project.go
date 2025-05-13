package models

import (
	"time" // Added import for time.Time

	"gorm.io/gorm"
)

// Project represents a project in the system
type Project struct {
	ID          int            `json:"id"` // JSON tag for ID
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	SecretKey   string         `gorm:"type:varchar(255)" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`           // JSON tag for CreatedAt
	UpdatedAt   time.Time      `json:"updatedAt"`           // JSON tag for UpdatedAt
	DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty"` // JSON tag for DeletedAt

	ReminderSchedules []ReminderSchedule `gorm:"foreignKey:ProjectID" json:"reminder_schedules,omitempty"`
	TotalReminders    int                `gorm:"-" json:"totalReminders"` // gorm:"-": This tag tells GORM (the ORM you're using) to ignore this field during database operations, meaning it won't try to find a corresponding column in the
}
