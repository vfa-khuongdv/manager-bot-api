package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type RecentScanResult struct {
	ID          string `json:"id"`
	ConfigName  string `json:"configName"`
	ConfigID    string `json:"configId"`
	ProjectID   int    `json:"projectId"`
	ProjectName string `json:"projectName"`
	LastScan    string `json:"lastScan"`
	VulnCount   int    `json:"vulnCount"`
	Status      string `json:"status"`
}

type ICveScanLogRepository interface {
	Create(log *models.CveScanLog) (*models.CveScanLog, error)
	Update(log *models.CveScanLog) (*models.CveScanLog, error)
	GetByConfigID(configID string, paging *utils.Paging) ([]models.CveScanLog, int64, error)
	CreateVulnerability(vuln *models.Vulnerability) error
	DeleteVulnerabilitiesByScanLogID(scanLogID uint) error
	GetLatestByConfigIDs(configIDs []string) ([]models.CveScanLog, error)
	GetVulnerabilitiesByScanLogIDs(scanLogIDs []uint) (map[uint][]models.Vulnerability, error)
	GetRecentScans(limit int) ([]RecentScanResult, int64, error)
}

type CveScanLogRepository struct {
	db *gorm.DB
}

func NewCveScanLogRepository(db *gorm.DB) *CveScanLogRepository {
	return &CveScanLogRepository{
		db: db,
	}
}

func (repo *CveScanLogRepository) Create(log *models.CveScanLog) (*models.CveScanLog, error) {
	if err := repo.db.Create(log).Error; err != nil {
		return nil, err
	}
	return log, nil
}

func (repo *CveScanLogRepository) Update(log *models.CveScanLog) (*models.CveScanLog, error) {
	if err := repo.db.Save(log).Error; err != nil {
		return nil, err
	}
	return log, nil
}

func (repo *CveScanLogRepository) GetByConfigID(configID string, paging *utils.Paging) ([]models.CveScanLog, int64, error) {
	var logs []models.CveScanLog

	q := repo.db.Model(&models.CveScanLog{}).Where("config_id = ?", configID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (paging.Page - 1) * paging.Limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(paging.Limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (repo *CveScanLogRepository) CreateVulnerability(vuln *models.Vulnerability) error {
	if err := repo.db.Create(vuln).Error; err != nil {
		logger.Warnf("CreateVulnerability failed: %v", err)
		return err
	}
	return nil
}

func (repo *CveScanLogRepository) DeleteVulnerabilitiesByScanLogID(scanLogID uint) error {
	if err := repo.db.Where("scan_log_id = ?", scanLogID).Delete(&models.Vulnerability{}).Error; err != nil {
		return err
	}
	return nil
}

func (repo *CveScanLogRepository) GetLatestByConfigIDs(configIDs []string) ([]models.CveScanLog, error) {
	if len(configIDs) == 0 {
		return nil, nil
	}

	var logs []models.CveScanLog
	subQuery := repo.db.Model(&models.CveScanLog{}).
		Select("MAX(id) as id").
		Where("config_id IN (?)", configIDs).
		Group("config_id")

	if err := repo.db.Where("id IN (?)", subQuery).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func (repo *CveScanLogRepository) GetVulnerabilitiesByScanLogIDs(scanLogIDs []uint) (map[uint][]models.Vulnerability, error) {
	if len(scanLogIDs) == 0 {
		return nil, nil
	}

	var vulns []models.Vulnerability
	if err := repo.db.Where("scan_log_id IN (?)", scanLogIDs).Find(&vulns).Error; err != nil {
		return nil, err
	}

	result := make(map[uint][]models.Vulnerability)
	for i := range vulns {
		result[vulns[i].ScanLogID] = append(result[vulns[i].ScanLogID], vulns[i])
	}

	return result, nil
}

func (repo *CveScanLogRepository) GetRecentScans(limit int) ([]RecentScanResult, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var results []RecentScanResult

	subQuery := repo.db.Model(&models.CveScanLog{}).
		Select("config_id, MAX(created_at) as last_created").
		Group("config_id")

	var total int64
	if err := repo.db.Model(&models.CveScanLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := repo.db.Table("(?) as latest", subQuery).
		Select("sl.id, cc.name as config_name, cc.id as config_id, cc.project_id, p.name as project_name, sl.created_at as last_scan, sl.vuln_found_count as vuln_count, sl.status").
		Joins("JOIN cve_scan_logs sl ON sl.config_id = latest.config_id AND sl.created_at = latest.last_created").
		Joins("JOIN cve_configs cc ON cc.id = sl.config_id").
		Joins("JOIN projects p ON p.id = cc.project_id").
		Order("sl.created_at DESC").
		Limit(limit).
		Find(&results).Error; err != nil {
		logger.Error("GetRecentScans query failed: ", err)
		return nil, 0, err
	}

	for i := range results {
		results[i].LastScan = results[i].LastScan
	}

	return results, total, nil
}
