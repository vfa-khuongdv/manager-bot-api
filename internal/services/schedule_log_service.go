package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type ScheduleLogService struct {
	repo repositories.IScheduleLogRepository
}

func NewScheduleLogService(repo repositories.IScheduleLogRepository) *ScheduleLogService {
	return &ScheduleLogService{repo: repo}
}

func (s *ScheduleLogService) GetDashboardData() (*models.DashboardData, error) {
	return s.repo.GetDashboardData()
}

// ListAll returns all run logs with optional filters and pagination (V2)
func (s *ScheduleLogService) ListAll(filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	return s.repo.ListAll(filters, paging)
}

// ListByProject returns run logs scoped to a project (V2)
func (s *ScheduleLogService) ListByProject(projectID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	return s.repo.ListByProject(projectID, filters, paging)
}

// ListBySchedule returns run logs scoped to a schedule (V2)
func (s *ScheduleLogService) ListBySchedule(scheduleID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	return s.repo.ListBySchedule(scheduleID, filters, paging)
}

// GetV2Summary returns aggregated V2 dashboard stats
func (s *ScheduleLogService) GetV2Summary() (*models.V2DashboardSummary, error) {
	return s.repo.GetV2Summary()
}
