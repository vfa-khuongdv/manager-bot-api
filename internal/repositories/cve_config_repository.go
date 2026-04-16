package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/gorm"
)

type ICveConfigRepository interface {
	GetAll(paging *utils.Paging) ([]models.CveConfig, int64, error)
	GetByProjectID(projectID uint, paging *utils.Paging) ([]models.CveConfig, int64, error)
	GetByID(id string) (*models.CveConfig, error)
	GetByUUID(id string, projectID uint) (*models.CveConfig, error)
	Create(config *models.CveConfig) (*models.CveConfig, error)
	Update(config *models.CveConfig) (*models.CveConfig, error)
	Delete(id string, projectID uint) error
	GetVulnerabilitiesByConfigID(configID string) ([]models.Vulnerability, int64, error)
	UpsertVulnerability(vuln *models.Vulnerability) error
	DeleteVulnerabilitiesByConfigID(configID string) error
}

type CveConfigRepository struct {
	db *gorm.DB
}

func NewCveConfigRepository(db *gorm.DB) *CveConfigRepository {
	return &CveConfigRepository{
		db: db,
	}
}

func (repo *CveConfigRepository) GetAll(paging *utils.Paging) ([]models.CveConfig, int64, error) {
	var configs []models.CveConfig

	q := repo.db.Model(&models.CveConfig{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (paging.Page - 1) * paging.Limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(paging.Limit).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

func (repo *CveConfigRepository) GetByProjectID(projectID uint, paging *utils.Paging) ([]models.CveConfig, int64, error) {
	var configs []models.CveConfig

	q := repo.db.Model(&models.CveConfig{}).Where("project_id = ?", projectID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (paging.Page - 1) * paging.Limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(paging.Limit).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

func (repo *CveConfigRepository) GetByID(id string) (*models.CveConfig, error) {
	var config models.CveConfig
	if err := repo.db.First(&config, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (repo *CveConfigRepository) GetByUUID(id string, projectID uint) (*models.CveConfig, error) {
	var config models.CveConfig
	if err := repo.db.First(&config, "id = ? AND project_id = ?", id, projectID).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (repo *CveConfigRepository) Create(config *models.CveConfig) (*models.CveConfig, error) {
	if err := repo.db.Create(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (repo *CveConfigRepository) Update(config *models.CveConfig) (*models.CveConfig, error) {
	if err := repo.db.Save(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (repo *CveConfigRepository) Delete(id string, projectID uint) error {
	if err := repo.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&models.CveConfig{}).Error; err != nil {
		return err
	}
	return nil
}

func (repo *CveConfigRepository) GetVulnerabilitiesByConfigID(configID string) ([]models.Vulnerability, int64, error) {
	var vulns []models.Vulnerability

	q := repo.db.Model(&models.Vulnerability{}).Where("config_id = ?", configID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("created_at DESC").Find(&vulns).Error; err != nil {
		return nil, 0, err
	}

	return vulns, total, nil
}

func (repo *CveConfigRepository) UpsertVulnerability(vuln *models.Vulnerability) error {
	return repo.db.Where("id = ? AND config_id = ?", vuln.ID, vuln.ConfigID).
		Assign(*vuln).
		FirstOrCreate(vuln).Error
}

func (repo *CveConfigRepository) DeleteVulnerabilitiesByConfigID(configID string) error {
	return repo.db.Where("config_id = ?", configID).Delete(&models.Vulnerability{}).Error
}
