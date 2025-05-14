package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
)

type IProjectService interface {
	GetAll() (*[]models.Project, error)
	GetByID(id uint) (*models.Project, error)
	Create(project *models.Project) (*models.Project, error)
	Update(project *models.Project) (*models.Project, error)
	Delete(id uint) error
	ValidateSecretKey(id uint, secretKey string) (bool, error)
}

type ProjectService struct {
	repo repositories.IProjectRepository
}

func NewProjectService(repo repositories.IProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (s *ProjectService) GetAll() (*[]models.Project, error) {
	projects, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectService) GetByID(id uint) (*models.Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Create(project *models.Project) (*models.Project, error) {
	createdProject, err := s.repo.Create(project)
	if err != nil {
		return nil, err
	}
	return createdProject, nil
}

func (s *ProjectService) Update(project *models.Project) (*models.Project, error) {
	updatedProject, err := s.repo.Update(project)
	if err != nil {
		return nil, err
	}
	return updatedProject, nil
}

func (s *ProjectService) Delete(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

// ValidateSecretKey validates if a project with the given ID has the provided secret key
// Parameters:
//   - id: The ID of the project to validate
//   - secretKey: The secret key to validate against the project
//
// Returns:
//   - bool: True if the secret key is valid, false otherwise
//   - error: An error if the project wasn't found or any other database error
func (s *ProjectService) ValidateSecretKey(id uint, secretKey string) (bool, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return false, err
	}

	// Validate the secret key
	return project.SecretKey == secretKey, nil
}
