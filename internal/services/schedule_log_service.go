package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
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
