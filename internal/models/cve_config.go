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
	NotifyOnModerate     bool           `gorm:"default:false" json:"notifyOnModerate"`
	NotifyOnLow          bool           `gorm:"default:false" json:"notifyOnLow"`
	LastScan             *time.Time     `json:"lastScan,omitempty"`
	LastStatus           string         `gorm:"type:varchar(20);default:'no_scan'" json:"lastStatus"`
	VulnerabilitiesFound int            `gorm:"default:0" json:"vulnerabilitiesFound"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (CveConfig) TableName() string {
	return "cve_configs"
}
