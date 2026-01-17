package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

type IScheduleLogRepository interface {
	GetDashboardData() (*models.DashboardData, error)
}

type ScheduleLogRepository struct {
	db *gorm.DB
}

func NewScheduleLogRepository(db *gorm.DB) *ScheduleLogRepository {
	return &ScheduleLogRepository{db: db}
}

func (r *ScheduleLogRepository) GetDashboardData() (*models.DashboardData, error) {
	var data models.DashboardData

	// Count total projects
	if err := r.db.Model(&models.Project{}).Count(&data.TotalProjects).Error; err != nil {
		return nil, err
	}

	// Count total schedules
	if err := r.db.Model(&models.ReminderSchedule{}).Count(&data.TotalSchedules).Error; err != nil {
		return nil, err
	}

	// Get project summaries
	var summaries []models.ProjectSummary
	query := `
		SELECT 
			p.id as project_id, 
			p.name as project_name,
			SUM(CASE WHEN sl.status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN sl.status = 'error' THEN 1 ELSE 0 END) as error_count
		FROM projects p
		LEFT JOIN schedule_logs sl ON p.id = sl.project_id
		WHERE p.deleted_at IS NULL
		GROUP BY p.id, p.name
	`

	if err := r.db.Raw(query).Scan(&summaries).Error; err != nil {
		return nil, err
	}
	data.ProjectSummaries = summaries

	// Get trend data for the past 7 days
	// If data for a day is missing, it should show 0 executions
	var trends []models.TrendDataPoint
	trendQuery := `
		WITH RECURSIVE date_series AS (
			SELECT CURDATE() - INTERVAL 6 DAY AS log_date
			UNION ALL
			SELECT log_date + INTERVAL 1 DAY
			FROM date_series
			WHERE log_date + INTERVAL 1 DAY <= CURDATE()
		)
		SELECT 
			DATE_FORMAT(ds.log_date, '%a') AS day,
			IFNULL(SUM(CASE WHEN DATE(sl.created_at) = ds.log_date THEN 1 ELSE 0 END), 0) AS executions
		FROM date_series ds
		LEFT JOIN schedule_logs sl ON DATE(sl.created_at) = ds.log_date
		GROUP BY ds.log_date
		ORDER BY ds.log_date;
	`

	if err := r.db.Raw(trendQuery).Scan(&trends).Error; err != nil {
		return nil, err
	}
	data.TrendData = trends

	return &data, nil
}
