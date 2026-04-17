package models

import (
	"time"

	"gorm.io/gorm"
)

type CveConfig struct {
	ID                   string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	ProjectID            int            `gorm:"type:int;not null;index" json:"projectId"`
	Name                 string         `gorm:"type:varchar(255);not null" json:"name"`
	RepoUrl              string         `gorm:"type:text" json:"repoUrl"`
	Languages            string         `gorm:"type:text;not null" json:"languages"`
	Cron                 string         `gorm:"type:varchar(50);not null" json:"cron"`
	Status               string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	ApiKey               string         `gorm:"type:varchar(255)" json:"-"`
	BotID                *int           `gorm:"type:int" json:"botId,omitempty"`
	NotifyOnSuccess      bool           `gorm:"default:false" json:"notifyOnSuccess"`
	NotifyOnFailure      bool           `gorm:"default:true" json:"notifyOnFailure"`
	NotifyRoomId         string         `gorm:"type:varchar(255)" json:"notifyRoomId,omitempty"`
	NotifyOnCritical     bool           `gorm:"default:true" json:"notifyOnCritical"`
	NotifyOnHigh         bool           `gorm:"default:true" json:"notifyOnHigh"`
	NotifyOnMedium       bool           `gorm:"default:false" json:"notifyOnMedium"`
	NotifyOnLow          bool           `gorm:"default:false" json:"notifyOnLow"`
	LastScan             *time.Time     `json:"lastScan,omitempty"`
	LastStatus           string         `gorm:"type:varchar(20);default:'no_scan'" json:"lastStatus"`
	VulnerabilitiesFound int            `gorm:"default:0" json:"vulnerabilitiesFound"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

type Vulnerability struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ScanLogID    uint      `gorm:"not null;index:idx_vuln_scan_log_id" json:"scanLogId"`
	ConfigID     string    `gorm:"type:varchar(36);not null;index" json:"configId"`
	CVEID        string    `gorm:"type:varchar(50);not null;index" json:"cveId"`
	Severity     string    `gorm:"type:varchar(20);not null" json:"severity"`
	Package      string    `gorm:"type:varchar(255);not null" json:"package"`
	Version      string    `gorm:"type:varchar(100);not null" json:"version"`
	Summary      string    `gorm:"type:text" json:"summary,omitempty"`
	Score        float64   `gorm:"type:decimal(5,2)" json:"score,omitempty"`
	ReferenceURL string    `gorm:"type:varchar(500)" json:"referenceUrl,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (CveConfig) TableName() string {
	return "cve_configs"
}

func (Vulnerability) TableName() string {
	return "vulnerabilities"
}
