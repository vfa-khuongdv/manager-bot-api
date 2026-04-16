package models

import (
	"time"

	"gorm.io/gorm"
)

type CveScanLog struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConfigID       string         `gorm:"type:varchar(36);not null;index:idx_log_config_id" json:"configId"`
	ProjectID      uint           `gorm:"not null;index:idx_log_project_id" json:"projectId"`
	Status         string         `gorm:"type:varchar(20);not null" json:"status"`
	VulnFoundCount int            `gorm:"default:0" json:"vulnFoundCount"`
	ErrorMessage   string         `gorm:"type:text" json:"errorMessage,omitempty"`
	StartedAt      time.Time      `json:"startedAt"`
	FinishedAt     *time.Time     `json:"finishedAt,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (CveScanLog) TableName() string {
	return "cve_scan_logs"
}
