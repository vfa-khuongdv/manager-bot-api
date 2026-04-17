package repositories

import (
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/gorm"
)

type ScheduleAnalysisRaw struct {
	ScheduleID  uint
	LastRun     *time.Time
	LastStatus  string
	TotalRuns   int64
	SuccessRuns int64
	FailedRuns  int64
}

type IScheduleLogRepository interface {
	GetDashboardData() (*models.DashboardData, error)
	ListAll(filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error)
	ListByProject(projectID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error)
	ListBySchedule(scheduleID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error)
	GetV2Summary() (*models.V2DashboardSummary, error)
	GetAnalysisByProject(projectID uint) ([]ScheduleAnalysisRaw, error)
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

// buildLogQuery builds a base query for schedule_logs joined with projects and schedules
func (r *ScheduleLogRepository) buildLogQuery(filters map[string]interface{}) *gorm.DB {
	q := r.db.Table("schedule_logs sl").
		Select("sl.id, sl.schedule_id, rs.name as name, p.name as project_name, sl.status, sl.created_at as timestamp, sl.error_message as message").
		Joins("LEFT JOIN projects p ON p.id = sl.project_id").
		Joins("LEFT JOIN reminder_schedules rs ON rs.id = sl.schedule_id").
		Where("sl.deleted_at IS NULL")

	if status, ok := filters["status"]; ok && status != "" {
		q = q.Where("sl.status = ?", status)
	}
	if from, ok := filters["from"]; ok && from != "" {
		q = q.Where("DATE(sl.created_at) >= ?", from)
	}
	if to, ok := filters["to"]; ok && to != "" {
		q = q.Where("DATE(sl.created_at) <= ?", to)
	}
	return q
}

func scanLogs(q *gorm.DB, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (paging.Page - 1) * paging.Limit
	var logs []models.RunLogV2
	if err := q.Order("sl.created_at DESC").Offset(offset).Limit(paging.Limit).Scan(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ListAll returns all run logs with optional filters
func (r *ScheduleLogRepository) ListAll(filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	return scanLogs(r.buildLogQuery(filters), paging)
}

// ListByProject returns run logs scoped to a project
func (r *ScheduleLogRepository) ListByProject(projectID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	q := r.buildLogQuery(filters).Where("sl.project_id = ?", projectID)
	return scanLogs(q, paging)
}

// ListBySchedule returns run logs scoped to a schedule
func (r *ScheduleLogRepository) ListBySchedule(scheduleID uint, filters map[string]interface{}, paging *utils.Paging) ([]models.RunLogV2, int64, error) {
	q := r.buildLogQuery(filters).Where("sl.schedule_id = ?", scheduleID)
	return scanLogs(q, paging)
}

// GetV2Summary returns aggregated stats for V2 dashboard
func (r *ScheduleLogRepository) GetV2Summary() (*models.V2DashboardSummary, error) {
	var summary models.V2DashboardSummary

	if err := r.db.Model(&models.Project{}).Where("status = ?", "active").Count(&summary.ActiveProjects).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.Project{}).Where("status = ?", "inactive").Count(&summary.InactiveProjects).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.ReminderSchedule{}).Count(&summary.TotalSchedules).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.ReminderSchedule{}).Where("active = ?", true).Count(&summary.ActiveSchedules).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.ScheduleLog{}).Where("status = ?", "success").Count(&summary.SuccessRuns).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.ScheduleLog{}).Where("status != ?", "success").Count(&summary.FailedRuns).Error; err != nil {
		return nil, err
	}

	total := summary.SuccessRuns + summary.FailedRuns
	if total > 0 {
		summary.SuccessRate = float64(summary.SuccessRuns) / float64(total) * 100
	}

	if err := r.db.Model(&models.CveConfig{}).Count(&summary.TotalCveConfigs).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.CveConfig{}).Where("status = ?", "active").Count(&summary.ActiveCveMonitoring).Error; err != nil {
		return nil, err
	}

	subQuery := r.db.Model(&models.CveScanLog{}).
		Select("config_id, MAX(created_at) as last_scan").
		Where("config_id IN (?)", r.db.Model(&models.CveConfig{}).Where("status = ?", "active").Select("id")).
		Group("config_id")

	var vulnSummary struct {
		TotalVulns  int64
		SecureCount int64
	}

	if err := r.db.Table("(?) as latest", subQuery).
		Select("COALESCE(SUM(sl.vuln_found_count), 0) as total_vulns, COALESCE(SUM(CASE WHEN sl.status = 'success' AND sl.vuln_found_count = 0 THEN 1 ELSE 0 END), 0) as secure_count").
		Joins("JOIN cve_scan_logs sl ON sl.config_id = latest.config_id AND sl.created_at = latest.last_scan").
		Scan(&vulnSummary).Error; err != nil {
		return nil, err
	}
	summary.TotalVulnerabilities = vulnSummary.TotalVulns
	summary.SecureConfigs = vulnSummary.SecureCount

	type VulnCount struct {
		Severity string
		Count    int64
	}
	var severityCounts []VulnCount

	subQuery2 := r.db.Model(&models.CveScanLog{}).
		Select("config_id, MAX(created_at) as last_scan").
		Where("config_id IN (?)", r.db.Model(&models.CveConfig{}).Where("status = ?", "active").Select("id")).
		Group("config_id")

	if err := r.db.Table("(?) as latest", subQuery2).
		Select("vuln.severity, COUNT(*) as count").
		Joins("JOIN cve_scan_logs sl ON sl.config_id = latest.config_id AND sl.created_at = latest.last_scan").
		Joins("JOIN vulnerabilities vuln ON vuln.scan_log_id = sl.id").
		Group("vuln.severity").
		Scan(&severityCounts).Error; err != nil {
		return nil, err
	}

	for _, v := range severityCounts {
		switch v.Severity {
		case "CRITICAL":
			summary.CriticalVulns = v.Count
		case "HIGH":
			summary.HighVulns = v.Count
		case "MODERATE":
			summary.ModerateVulns = v.Count
		case "LOW":
			summary.LowVulns = v.Count
		default:
			summary.ModerateVulns += v.Count
		}
	}

	return &summary, nil
}

func (r *ScheduleLogRepository) GetAnalysisByProject(projectID uint) ([]ScheduleAnalysisRaw, error) {
	var results []ScheduleAnalysisRaw

	err := r.db.Model(&models.ScheduleLog{}).
		Select("schedule_id, MAX(created_at) as last_run, MAX(status) as last_status, COUNT(*) as total_runs, SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_runs, SUM(CASE WHEN status != 'success' THEN 1 ELSE 0 END) as failed_runs").
		Where("project_id = ?", projectID).
		Group("schedule_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for i := range results {
		results[i].SuccessRuns = results[i].TotalRuns - results[i].FailedRuns
		results[i].FailedRuns = results[i].FailedRuns
		results[i].TotalRuns = results[i].SuccessRuns + results[i].FailedRuns
	}

	return results, nil
}
