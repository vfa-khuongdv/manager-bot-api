package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	GetAll() (*[]models.Project, error)
	GetByID(id uint) (*models.Project, error)
	Create(project *models.Project) (*models.Project, error)
	Update(project *models.Project) (*models.Project, error)
	Delete(id uint) error
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (repo *ProjectRepository) GetAll() (*[]models.Project, error) {
	var projects []models.Project
	if err := repo.db.Preload("ReminderSchedules").Find(&projects).Error; err != nil {
		return nil, err
	}

	for i := range projects {
		projects[i].TotalReminders = len(projects[i].ReminderSchedules)
	}

	return &projects, nil
}

func (repo *ProjectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	if err := repo.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (repo *ProjectRepository) Create(project *models.Project) (*models.Project, error) {
	if err := repo.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

func (repo *ProjectRepository) Update(project *models.Project) (*models.Project, error) {
	if err := repo.db.Save(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

func (repo *ProjectRepository) Delete(id uint) error {
	if err := repo.db.Delete(&models.Project{}, id).Error; err != nil {
		return err
	}
	return nil
}
