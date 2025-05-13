package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type IReminderScheduleService interface {
	GetAll(paging *utils.Paging) ([]models.ReminderSchedule, error)
	GetByID(id uint) (*models.ReminderSchedule, error)
	GetByProjectID(projectID uint) ([]models.ReminderSchedule, error)
	Create(schedule *models.ReminderSchedule) error
	Update(schedule *models.ReminderSchedule) error
	Delete(id uint) error
	GetActiveSchedules() ([]models.ReminderSchedule, error)
	ToggleActiveStatus(id uint, active bool) error
}

type ReminderScheduleService struct {
	repo *repositories.ReminderScheduleRepository
}

// NewReminderScheduleService creates a new instance of ReminderScheduleService
// Parameters:
//   - repo: pointer to ReminderScheduleRepository for database operations
//
// Returns:
//   - *ReminderScheduleService: new instance of the service
func NewReminderScheduleService(repo *repositories.ReminderScheduleRepository) *ReminderScheduleService {
	return &ReminderScheduleService{
		repo: repo,
	}
}

// GetAll retrieves all reminder schedules with pagination
// Parameters:
//   - paging: pointer to utils.Paging for pagination control
//
// Returns:
//   - []models.ReminderSchedule: slice of reminder schedules
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) GetAll(paging *utils.Paging) ([]models.ReminderSchedule, error) {
	schedules, err := service.repo.GetAll(paging)
	if err != nil {
		return nil, errors.New(errors.ErrDatabaseQuery, err.Error())
	}
	return schedules, nil
}

// GetByID retrieves a reminder schedule by its ID
// Parameters:
//   - id: the unique identifier of the schedule to retrieve
//
// Returns:
//   - *models.ReminderSchedule: pointer to the retrieved schedule
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) GetByID(id uint) (*models.ReminderSchedule, error) {
	schedule, err := service.repo.GetByID(id)
	if err != nil {
		return nil, errors.New(errors.ErrDatabaseQuery, err.Error())
	}
	return schedule, nil
}

// GetByProjectID retrieves all reminder schedules for a specific project
// Parameters:
//   - projectID: the ID of the project to get schedules for
//
// Returns:
//   - []models.ReminderSchedule: slice of reminder schedules
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) GetByProjectID(projectID uint) ([]models.ReminderSchedule, error) {
	schedules, err := service.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, errors.New(errors.ErrDatabaseQuery, err.Error())
	}
	return schedules, nil
}

// Create saves a new reminder schedule
// Parameters:
//   - schedule: pointer to the schedule model to create
//
// Returns:
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) Create(schedule *models.ReminderSchedule) error {
	if err := service.repo.Create(schedule); err != nil {
		return errors.New(errors.ErrDatabaseInsert, err.Error())
	}
	return nil
}

// Update modifies an existing reminder schedule
// Parameters:
//   - schedule: pointer to the schedule model with updated data
//
// Returns:
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) Update(schedule *models.ReminderSchedule) error {
	// Check if schedule exists
	existingSchedule, err := service.repo.GetByID(schedule.ID)
	if err != nil {
		return errors.New(errors.ErrDatabaseQuery, err.Error())
	}

	if existingSchedule == nil {
		return errors.New(errors.ErrServerInternal, "reminder schedule not found")
	}

	if err := service.repo.Update(schedule); err != nil {
		return errors.New(errors.ErrDatabaseUpdate, err.Error())
	}
	return nil
}

// Delete removes a reminder schedule
// Parameters:
//   - id: the ID of the schedule to delete
//
// Returns:
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) Delete(id uint) error {
	schedule, err := service.repo.GetByID(id)
	if err != nil {
		return errors.New(errors.ErrDatabaseQuery, err.Error())
	}

	if schedule == nil {
		return errors.New(errors.ErrResourceNotFound, "reminder schedule not found")
	}

	if err := service.repo.Delete(schedule); err != nil {
		return errors.New(errors.ErrDatabaseDelete, err.Error())
	}
	return nil
}

// GetActiveSchedules retrieves all active reminder schedules
// Returns:
//   - []models.ReminderSchedule: slice of active reminder schedules
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) GetActiveSchedules() ([]models.ReminderSchedule, error) {
	schedules, err := service.repo.GetActiveSchedules()
	if err != nil {
		return nil, errors.New(errors.ErrDatabaseQuery, err.Error())
	}
	return schedules, nil
}

// ToggleActiveStatus updates the active status of a reminder schedule
// Parameters:
//   - id: the ID of the schedule to update
//   - active: the new active status
//
// Returns:
//   - error: nil if successful, wrapped error otherwise
func (service *ReminderScheduleService) ToggleActiveStatus(id uint, active bool) error {
	// Check if schedule exists
	_, err := service.repo.GetByID(id)
	if err != nil {
		return errors.New(errors.ErrDatabaseQuery, err.Error())
	}

	if err := service.repo.UpdateActiveStatus(id, active); err != nil {
		return errors.New(errors.ErrDatabaseUpdate, err.Error())
	}
	return nil
}
