package models

import (
	"time"

	"gorm.io/gorm"
)

type ScheduleLog struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ProjectID    uint           `json:"projectId" gorm:"column:project_id;not null;index:idx_log_project_id"`
	ScheduleID   uint           `json:"scheduleId" gorm:"column:schedule_id;not null;index:idx_log_schedule_id"`
	Status       string         `json:"status" gorm:"column:status;type:varchar(50);not null"`
	ErrorMessage string         `json:"errorMessage" gorm:"column:error_message;type:text"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`

	Project  Project          `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Schedule ReminderSchedule `json:"schedule,omitempty" gorm:"foreignKey:ScheduleID"`
}

type ProjectSummary struct {
	ProjectID    uint   `json:"projectId"`
	ProjectName  string `json:"projectName"`
	SuccessCount int64  `json:"successCount"`
	ErrorCount   int64  `json:"errorCount"`
}

type TrendDataPoint struct {
	Day        string `json:"day"`
	Executions int64  `json:"executions"`
}

type DashboardData struct {
	TotalProjects    int64            `json:"totalProjects"`
	TotalSchedules   int64            `json:"totalSchedules"`
	ProjectSummaries []ProjectSummary `json:"projectSummaries"`
	TrendData        []TrendDataPoint `json:"trends"`
}

// V2DashboardSummary provides aggregated stats for V2 dashboard
type V2DashboardSummary struct {
	ActiveProjects   int64   `json:"activeProjects"`
	InactiveProjects int64   `json:"inactiveProjects"`
	TotalSchedules   int64   `json:"totalSchedules"`
	ActiveSchedules  int64   `json:"activeSchedules"`
	SuccessRuns      int64   `json:"successRuns"`
	FailedRuns       int64   `json:"failedRuns"`
	SuccessRate      float64 `json:"successRate"`
}

// RunLogV2 is the V2 API response shape for a run log entry
type RunLogV2 struct {
	ID          uint   `json:"id"`
	ScheduleID  uint   `json:"scheduleId"`
	Name        string `json:"name"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	Message     string `json:"message"`
}
