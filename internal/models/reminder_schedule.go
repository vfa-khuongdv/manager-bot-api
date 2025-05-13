package models

import (
	"time"

	"gorm.io/gorm"
)

// ReminderSchedule represents a scheduled reminder for a project
type ReminderSchedule struct {
	ID             uint           `json:"id"` // JSON tag for ID
	ProjectID      uint           `json:"projectId" gorm:"column:project_id;not null;index:idx_reminder_project_id"`
	CronExpression string         `json:"cronExpression" gorm:"column:cron_expression;type:varchar(255);not null"`
	ChatworkRoomID string         `json:"chatworkRoomId" gorm:"column:chatwork_room_id;type:varchar(255);not null"`
	ChatworkToken  string         `json:"chatworkToken" gorm:"column:chatwork_token;type:varchar(255);not null"`
	Message        string         `json:"message" gorm:"column:message;type:text"`
	Active         bool           `json:"active" gorm:"column:active;default:true"`
	CreatedAt      time.Time      `json:"createdAt"`           // JSON tag for CreatedAt
	UpdatedAt      time.Time      `json:"updatedAt"`           // JSON tag for UpdatedAt
	DeletedAt      gorm.DeletedAt `json:"deletedAt,omitempty"` // JSON tag for DeletedAt
	Project        Project        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
}
