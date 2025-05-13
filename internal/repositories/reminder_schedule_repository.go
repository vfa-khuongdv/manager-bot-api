package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/gorm"
)

type IReminderScheduleRepository interface {
	GetAll(paging *utils.Paging) ([]models.ReminderSchedule, error)
	GetByID(id uint) (*models.ReminderSchedule, error)
	GetByProjectID(projectID uint) ([]models.ReminderSchedule, error)
	Create(schedule *models.ReminderSchedule) error
	Update(schedule *models.ReminderSchedule) error
	Delete(schedule *models.ReminderSchedule) error
	GetActiveSchedules() ([]models.ReminderSchedule, error)
	UpdateActiveStatus(id uint, active bool) error
}

type ReminderScheduleRepository struct {
	db *gorm.DB
}

// NewReminderScheduleRepository creates a new instance of ReminderScheduleRepository
// Parameters:
//   - db: pointer to the gorm.DB instance for database operations
//
// Returns:
//   - *ReminderScheduleRepository: pointer to the newly created repository
func NewReminderScheduleRepository(db *gorm.DB) *ReminderScheduleRepository {
	return &ReminderScheduleRepository{
		db: db,
	}
}

// GetAll retrieves all reminder schedules with pagination support
// Parameters:
//   - paging: pointer to Paging structure for pagination control
//
// Returns:
//   - []models.ReminderSchedule: slice of reminder schedule models
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) GetAll(paging *utils.Paging) ([]models.ReminderSchedule, error) {
	var schedules []models.ReminderSchedule

	query := repo.db.Model(&models.ReminderSchedule{})
	query = utils.ApplyPaging(query, paging)

	err := query.Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// GetByID retrieves a reminder schedule by its ID
// Parameters:
//   - id: the unique identifier of the reminder schedule
//
// Returns:
//   - *models.ReminderSchedule: pointer to the found schedule, nil if not found
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) GetByID(id uint) (*models.ReminderSchedule, error) {
	var schedule models.ReminderSchedule
	if err := repo.db.First(&schedule, id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

// GetByProjectID retrieves all reminder schedules for a specific project
// Parameters:
//   - projectID: the ID of the project to get schedules for
//
// Returns:
//   - []models.ReminderSchedule: slice of reminder schedule models
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) GetByProjectID(projectID uint) ([]models.ReminderSchedule, error) {
	var schedules []models.ReminderSchedule
	if err := repo.db.Where("project_id = ?", projectID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

// Create saves a new reminder schedule to the database
// Parameters:
//   - schedule: pointer to the ReminderSchedule model to be created
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) Create(schedule *models.ReminderSchedule) error {
	return repo.db.Create(schedule).Error
}

// Update modifies an existing reminder schedule in the database
// Parameters:
//   - schedule: pointer to the ReminderSchedule model to be updated
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) Update(schedule *models.ReminderSchedule) error {
	return repo.db.Save(schedule).Error
}

// Delete removes a reminder schedule from the database
// Parameters:
//   - schedule: pointer to the ReminderSchedule model to be deleted
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) Delete(schedule *models.ReminderSchedule) error {
	return repo.db.Delete(schedule).Error
}

// GetActiveSchedules retrieves all active reminder schedules
// Returns:
//   - []models.ReminderSchedule: slice of active reminder schedule models
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) GetActiveSchedules() ([]models.ReminderSchedule, error) {
	var schedules []models.ReminderSchedule
	if err := repo.db.Where("active = ?", true).Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

// UpdateActiveStatus updates the active status of a reminder schedule
// Parameters:
//   - id: the ID of the reminder schedule to update
//   - active: the new active status (true or false)
//
// Returns:
//   - error: nil if successful, error otherwise
func (repo *ReminderScheduleRepository) UpdateActiveStatus(id uint, active bool) error {
	return repo.db.Model(&models.ReminderSchedule{}).Where("id = ?", id).Update("active", active).Error
}
