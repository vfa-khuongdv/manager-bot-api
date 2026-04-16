package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type ICveConfigService interface {
	GetByProjectID(projectID uint, paging *utils.Paging) ([]models.CveConfig, int64, error)
	GetByID(id string, projectID uint) (*models.CveConfig, error)
	Create(projectID uint, input *CveConfigInput) (*models.CveConfig, error)
	Update(id string, projectID uint, input *CveConfigUpdateInput) (*models.CveConfig, error)
	Delete(id string, projectID uint) error
	Toggle(id string, projectID uint) (*models.CveConfig, error)
	TriggerScan(id string, projectID uint) error
	GetVulnerabilities(configID string, projectID uint) ([]models.Vulnerability, int64, error)
	TestScan(languages string) ([]models.Vulnerability, error)
	GetScanLogs(configID string, projectID uint, paging *utils.Paging) ([]models.CveScanLog, int64, error)
	GetAnalysisByProject(projectID uint) ([]CveAnalysis, error)
}

type CveAnalysis struct {
	ConfigID        string                 `json:"configId"`
	ConfigName      string                 `json:"configName"`
	ConfigStatus    string                 `json:"configStatus"`
	LastScan        *time.Time             `json:"lastScan"`
	LastStatus      string                 `json:"lastStatus"`
	Vulnerabilities []models.Vulnerability `json:"vulnerabilities"`
}

type CveConfigService struct {
	repo            repositories.ICveConfigRepository
	logRepo         repositories.ICveScanLogRepository
	chatworkSvc     *ChatworkService
	chatworkBotRepo repositories.IChatworkBotRepository
}

func NewCveConfigService(repo repositories.ICveConfigRepository, logRepo repositories.ICveScanLogRepository, botRepo repositories.IChatworkBotRepository) *CveConfigService {
	return &CveConfigService{
		repo:            repo,
		logRepo:         logRepo,
		chatworkSvc:     NewChatworkService(),
		chatworkBotRepo: botRepo,
	}
}

type CveConfigForCron struct {
	ID              string
	ProjectID       int
	Name            string
	Cron            string
	NotifyOnSuccess bool
	NotifyOnFailure bool
	NotifyRoomId    string
	ApiKey          string
}

func (s *CveConfigService) GetAllForCron() ([]CveConfigForCron, error) {
	configs, _, err := s.repo.GetAll(&utils.Paging{Page: 1, Limit: 1000})
	if err != nil {
		return nil, err
	}

	var result []CveConfigForCron
	for _, cfg := range configs {
		if cfg.Status != "active" || cfg.Cron == "" {
			continue
		}
		result = append(result, CveConfigForCron{
			ID:              cfg.ID,
			ProjectID:       cfg.ProjectID,
			Name:            cfg.Name,
			Cron:            cfg.Cron,
			NotifyOnSuccess: cfg.NotifyOnSuccess,
			NotifyOnFailure: cfg.NotifyOnFailure,
			NotifyRoomId:    cfg.NotifyRoomId,
			ApiKey:          cfg.ApiKey,
		})
	}
	return result, nil
}

type CveConfigInput struct {
	Name            string `json:"name" binding:"required"`
	RepoUrl         string `json:"repoUrl"`
	Languages       string `json:"languages" binding:"required"`
	Cron            string `json:"cron" binding:"required"`
	Status          string `json:"status"`
	ApiKey          string `json:"apiKey"`
	BotID           *int   `json:"botId"`
	NotifyOnSuccess bool   `json:"notifyOnSuccess"`
	NotifyOnFailure bool   `json:"notifyOnFailure"`
	NotifyRoomId    string `json:"notifyRoomId"`
}

type CveConfigUpdateInput struct {
	Name            *string `json:"name"`
	RepoUrl         *string `json:"repoUrl"`
	Languages       *string `json:"languages"`
	Cron            *string `json:"cron"`
	Status          *string `json:"status"`
	ApiKey          *string `json:"apiKey"`
	BotID           *int    `json:"botId"`
	NotifyOnSuccess *bool   `json:"notifyOnSuccess"`
	NotifyOnFailure *bool   `json:"notifyOnFailure"`
	NotifyRoomId    *string `json:"notifyRoomId"`
}

func (s *CveConfigService) GetByProjectID(projectID uint, paging *utils.Paging) ([]models.CveConfig, int64, error) {
	return s.repo.GetByProjectID(projectID, paging)
}

func (s *CveConfigService) GetByID(id string, projectID uint) (*models.CveConfig, error) {
	return s.repo.GetByUUID(id, projectID)
}

func (s *CveConfigService) Create(projectID uint, input *CveConfigInput) (*models.CveConfig, error) {
	if input.Name == "" || input.Languages == "" || input.Cron == "" {
		return nil, fmt.Errorf("name, languages, and cron are required")
	}

	status := input.Status
	if status == "" {
		status = "active"
	}

	config := &models.CveConfig{
		ID:              generateUUID(),
		ProjectID:       int(projectID),
		Name:            input.Name,
		RepoUrl:         input.RepoUrl,
		Languages:       input.Languages,
		Cron:            input.Cron,
		Status:          status,
		ApiKey:          input.ApiKey,
		BotID:           input.BotID,
		NotifyOnSuccess: input.NotifyOnSuccess,
		NotifyOnFailure: input.NotifyOnFailure,
		NotifyRoomId:    input.NotifyRoomId,
	}

	return s.repo.Create(config)
}

func (s *CveConfigService) Update(id string, projectID uint, input *CveConfigUpdateInput) (*models.CveConfig, error) {
	config, err := s.repo.GetByUUID(id, projectID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		config.Name = *input.Name
	}
	if input.RepoUrl != nil {
		config.RepoUrl = *input.RepoUrl
	}
	if input.Languages != nil {
		config.Languages = *input.Languages
	}
	if input.Cron != nil {
		config.Cron = *input.Cron
	}
	if input.Status != nil {
		config.Status = *input.Status
	}
	if input.BotID != nil {
		config.BotID = input.BotID
	}
	if input.ApiKey != nil && *input.ApiKey != "" {
		config.ApiKey = *input.ApiKey
	}
	if input.NotifyOnSuccess != nil {
		config.NotifyOnSuccess = *input.NotifyOnSuccess
	}
	if input.NotifyOnFailure != nil {
		config.NotifyOnFailure = *input.NotifyOnFailure
	}
	if input.NotifyRoomId != nil && *input.NotifyRoomId != "" {
		config.NotifyRoomId = *input.NotifyRoomId
	}

	return s.repo.Update(config)
}

func (s *CveConfigService) Delete(id string, projectID uint) error {
	if err := s.repo.DeleteVulnerabilitiesByConfigID(id); err != nil {
		logger.Warnf("Failed to delete vulnerabilities: %v", err)
	}
	return s.repo.Delete(id, projectID)
}

func (s *CveConfigService) Toggle(id string, projectID uint) (*models.CveConfig, error) {
	config, err := s.repo.GetByUUID(id, projectID)
	if err != nil {
		return nil, err
	}

	if config.Status == "active" {
		config.Status = "paused"
	} else {
		config.Status = "active"
	}

	return s.repo.Update(config)
}

func (s *CveConfigService) TriggerScan(id string, projectID uint) error {
	config, err := s.repo.GetByUUID(id, projectID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	startedAt := time.Now()
	scanLog := &models.CveScanLog{
		ConfigID:  id,
		ProjectID: projectID,
		Status:    "running",
		StartedAt: startedAt,
	}

	createdLog, err := s.logRepo.Create(scanLog)
	if err != nil {
		logger.Warnf("Failed to create scan log: %v", err)
	}

	vulns, err := s.scanConfig(config)
	finishedAt := time.Now()

	if err != nil {
		config.LastStatus = "failed"
		config.LastScan = &finishedAt
		s.repo.Update(config)

		createdLog.Status = "failed"
		createdLog.ErrorMessage = err.Error()
		createdLog.FinishedAt = &finishedAt
		s.logRepo.Update(createdLog)

		return fmt.Errorf("scan failed: %w", err)
	}

	config.LastStatus = "success"
	config.LastScan = &finishedAt
	config.VulnerabilitiesFound = len(vulns)

	for i := range vulns {
		vuln := &vulns[i]
		vuln.ConfigID = id
		vuln.ScanLogID = createdLog.ID
		if err := s.logRepo.CreateVulnerability(vuln); err != nil {
			logger.Warnf("Failed to create vulnerability: %v", err)
		}
	}

	s.repo.Update(config)

	createdLog.Status = "success"
	createdLog.VulnFoundCount = len(vulns)
	createdLog.FinishedAt = &finishedAt
	s.logRepo.Update(createdLog)

	s.sendNotifications(config, vulns)

	return nil
}

func (s *CveConfigService) sendNotifications(config *models.CveConfig, vulns []models.Vulnerability) {
	if config.NotifyRoomId == "" {
		logger.Warn("[CVE] Notification skipped: no notifyRoomId")
		return
	}

	shouldNotify := (len(vulns) > 0 && config.NotifyOnFailure) || (len(vulns) == 0 && config.NotifyOnSuccess)
	if !shouldNotify {
		return
	}

	token := config.ApiKey
	if config.BotID != nil && token == "" {
		bot, err := s.chatworkBotRepo.GetByID(uint(*config.BotID))
		if err != nil {
			logger.Errorf("[CVE] Failed to get bot: %v", err)
			return
		}
		if bot != nil {
			token = bot.APIToken
		}
	}

	if token == "" {
		logger.Warn("[CVE] Notification skipped: no API key or bot")
		return
	}

	message := formatCVEMessage(config, vulns)
	logger.Infof("[CVE] Sending message to room %s: %s", config.NotifyRoomId, message)

	if err := s.chatworkSvc.SendMessage(token, config.NotifyRoomId, message); err != nil {
		logger.Errorf("[CVE] Failed to send notification: %v", err)
	}
}

func formatCVEMessage(config *models.CveConfig, vulns []models.Vulnerability) string {
	emoji := "✅"
	status := "No Vulnerabilities"
	if len(vulns) > 0 {
		emoji = "🚨"
		status = fmt.Sprintf("Found %d Vulnerabilities", len(vulns))
	}

	msg := fmt.Sprintf("[info][title]%s CVE Scan - %s[/title]", emoji, status)
	msg += fmt.Sprintf("\n⚙️ Config: %s", config.Name)
	msg += fmt.Sprintf("\n📦 Packages: %s", config.Languages)

	if len(vulns) > 0 {
		msg += "\n[hr]"
		msg += "\nVulnerabilities:"
		for i, v := range vulns {
			if i >= 5 {
				msg += fmt.Sprintf("\n... and %d more", len(vulns)-5)
				break
			}
			msg += fmt.Sprintf("\n• %s (%s) - ⭐ %.0f", v.CVEID, v.Severity, v.Score)
		}
	}
	msg += "\n[/info]"

	return msg
}

func (s *CveConfigService) GetVulnerabilities(configID string, projectID uint) ([]models.Vulnerability, int64, error) {
	_, err := s.repo.GetByUUID(configID, projectID)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.GetVulnerabilitiesByConfigID(configID)
}

func (s *CveConfigService) TestScan(languages string) ([]models.Vulnerability, error) {
	queries, err := parseLanguages(languages)
	if err != nil {
		return nil, err
	}

	if len(queries) == 0 {
		return nil, nil
	}

	osvResp, err := callOSVAPI(queries)
	if err != nil {
		logger.Warnf("CVE TestScan: callOSVAPI error: %v", err)
		return nil, err
	}

	var vulns []models.Vulnerability
	for i, result := range osvResp.Results {
		if result.Vulns == nil || len(result.Vulns) == 0 {
			continue
		}
		if i >= len(queries) {
			continue
		}
		pkg := queries[i]
		for _, v := range result.Vulns {
			detail := getVulnDetails(v.ID)
			severity := extractSeverity(detail)
			summary := v.Summary
			referenceURL := ""
			if detail.Summary != "" {
				summary = detail.Summary
			}

			if len(detail.References) > 0 {
				referenceURL = detail.References[0].URL
			}

			score := extractScore(detail)
			cveID := v.ID

			vuln := models.Vulnerability{
				CVEID:        cveID,
				Severity:     severity,
				Package:      pkg.Package.Name,
				Version:      pkg.Version,
				Summary:      summary,
				Score:        score,
				ReferenceURL: referenceURL,
			}
			vulns = append(vulns, vuln)
		}
	}

	return vulns, nil
}

func (s *CveConfigService) GetScanLogs(configID string, projectID uint, paging *utils.Paging) ([]models.CveScanLog, int64, error) {
	_, err := s.repo.GetByUUID(configID, projectID)
	if err != nil {
		return nil, 0, err
	}
	return s.logRepo.GetByConfigID(configID, paging)
}

func (s *CveConfigService) GetAnalysisByProject(projectID uint) ([]CveAnalysis, error) {
	configs, _, err := s.repo.GetByProjectID(projectID, &utils.Paging{Page: 1, Limit: 100})
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, nil
	}

	configIDs := make([]string, len(configs))
	configMap := make(map[string]*models.CveConfig)
	for i := range configs {
		configIDs[i] = configs[i].ID
		configMap[configs[i].ID] = &configs[i]
	}

	logs, err := s.logRepo.GetLatestByConfigIDs(configIDs)
	if err != nil {
		return nil, err
	}

	logger.Infof("CVE Analysis: found %d configs, %d logs", len(configs), len(logs))

	if len(logs) == 0 {
		var result []CveAnalysis
		for i := range configs {
			result = append(result, CveAnalysis{
				ConfigID:     configs[i].ID,
				ConfigName:   configs[i].Name,
				ConfigStatus: configs[i].Status,
				LastScan:     configs[i].LastScan,
				LastStatus:   configs[i].LastStatus,
			})
		}
		return result, nil
	}

	scanLogIDs := make([]uint, len(logs))
	logMap := make(map[uint]*models.CveScanLog)
	for i := range logs {
		scanLogIDs[i] = logs[i].ID
		logMap[logs[i].ID] = &logs[i]
	}

	vulnsMap, err := s.logRepo.GetVulnerabilitiesByScanLogIDs(scanLogIDs)
	if err != nil {
		return nil, err
	}

	logger.Infof("CVE Analysis: found %d scanLogs, %d vuln entries in map", len(scanLogIDs), len(vulnsMap))

	var result []CveAnalysis
	for i := range logs {
		log := logMap[logs[i].ID]
		config := configMap[log.ConfigID]

		analysis := CveAnalysis{
			ConfigID:     log.ConfigID,
			ConfigName:   config.Name,
			ConfigStatus: config.Status,
			LastScan:     &log.StartedAt,
			LastStatus:   log.Status,
		}

		if vulns, ok := vulnsMap[log.ID]; ok {
			analysis.Vulnerabilities = vulns
		}

		result = append(result, analysis)
	}

	for _, config := range configs {
		found := false
		for _, r := range result {
			if r.ConfigID == config.ID {
				found = true
				break
			}
		}
		if !found {
			result = append(result, CveAnalysis{
				ConfigID:     config.ID,
				ConfigName:   config.Name,
				ConfigStatus: config.Status,
				LastScan:     config.LastScan,
				LastStatus:   config.LastStatus,
			})
		}
	}

	return result, nil
}

func (s *CveConfigService) scanConfig(config *models.CveConfig) ([]models.Vulnerability, error) {
	queries, err := parseLanguages(config.Languages)
	if err != nil {
		return nil, err
	}

	if len(queries) == 0 {
		return nil, nil
	}

	osvResp, err := callOSVAPI(queries)
	if err != nil {
		return nil, err
	}

	var vulns []models.Vulnerability
	for i, result := range osvResp.Results {
		if result.Vulns == nil || len(result.Vulns) == 0 {
			continue
		}
		if i >= len(queries) {
			continue
		}
		pkg := queries[i]
		for _, v := range result.Vulns {
			details := getVulnDetails(v.ID)
			severity := extractSeverity(details)
			if severity == "" {
				severity = "medium"
			}
			summary := v.Summary
			refURL := ""
			if details.Summary != "" {
				summary = details.Summary
			}
			if len(details.References) > 0 {
				refURL = details.References[0].URL
			}

			score := extractScore(details)

			vulns = append(vulns, models.Vulnerability{
				CVEID:        v.ID,
				ConfigID:     config.ID,
				Severity:     severity,
				Package:      pkg.Package.Name,
				Version:      pkg.Version,
				Summary:      summary,
				Score:        score,
				ReferenceURL: refURL,
			})
		}
	}

	return vulns, nil
}

func parseLanguages(languages string) ([]OSVQuery, error) {
	var queries []OSVQuery
	if languages == "" {
		return queries, nil
	}

	parts := strings.Split(languages, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		idx := strings.Index(part, ":")
		if idx == -1 {
			logger.Warnf("parseLanguages: missing colon in %s", part)
			continue
		}

		ecosystem := strings.TrimSpace(part[:idx])
		rest := strings.TrimSpace(part[idx+1:])

		// Map common ecosystem aliases
		ecosystem = mapEcosystem(ecosystem)

		if !isValidEcosystem(ecosystem) {
			logger.Warnf("parseLanguages: invalid ecosystem %s", ecosystem)
			continue
		}

		pkgName := rest
		version := ""

		// Check if version is specified with @
		verIdx := strings.LastIndex(rest, "@")
		if verIdx != -1 {
			pkgName = rest[:verIdx]
			version = rest[verIdx+1:]
		}

		if pkgName == "" {
			logger.Warnf("parseLanguages: empty package name in %s", part)
			continue
		}

		queries = append(queries, OSVQuery{
			Package: OSPackage{
				Name:      pkgName,
				Ecosystem: ecosystem,
			},
			Version: version,
		})
	}

	return queries, nil
}

func isValidEcosystem(eco string) bool {
	valid := []string{"npm", "PyPI", "Maven", "Go", "crates.io", "NuGet", "RubyGems", "Packagist", "Pub", "SwiftPM"}
	for _, v := range valid {
		if eco == v {
			return true
		}
	}
	return false
}

func mapEcosystem(eco string) string {
	mapping := map[string]string{
		"python": "PyPI",
		"pip":    "PyPI",
		"node":   "npm",
		"nodejs": "npm",
		"java":   "Maven",
		"maven":  "Maven",
		"go":     "Go",
		"rust":   "crates.io",
		"dotnet": "NuGet",
		"nuget":  "NuGet",
		"ruby":   "RubyGems",
		"php":    "Packagist",
	}
	if mapped, ok := mapping[eco]; ok {
		return mapped
	}
	return eco
}

func callOSVAPI(queries []OSVQuery) (*OSVResponse, error) {
	url := "https://api.osv.dev/v1/querybatch"
	body := OSVBatchRequest{
		Queries: queries,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warnf("OSV API request failed: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("OSV API returned status %d", resp.StatusCode)
		return nil, fmt.Errorf("OSV API returned status %d", resp.StatusCode)
	}

	var osvResp OSVResponse
	if err := json.NewDecoder(resp.Body).Decode(&osvResp); err != nil {
		logger.Warnf("OSV API failed to decode: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &osvResp, nil
}

func (s *CveConfigService) sendNotification(config *models.CveConfig, vulns []models.Vulnerability, roomID string) {
	if config.ApiKey == "" || roomID == "" {
		return
	}

	message := formatVulnerabilityMessage(config, vulns)
	if err := s.chatworkSvc.SendMessage(config.ApiKey, roomID, message); err != nil {
		logger.Warnf("Failed to send notification: %v", err)
	}
}

func formatVulnerabilityMessage(config *models.CveConfig, vulns []models.Vulnerability) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[info][title]🚨 CVE Scan Results: %s[/title]\n", config.Name))
	sb.WriteString(fmt.Sprintf("📊 Total vulnerabilities found: %d\n\n", len(vulns)))

	bySeverity := make(map[string][]models.Vulnerability)
	for _, v := range vulns {
		bySeverity[v.Severity] = append(bySeverity[v.Severity], v)
	}

	for _, sev := range []string{"critical", "high", "medium", "low"} {
		if vulns, ok := bySeverity[sev]; ok {
			icon := map[string]string{"critical": "🔴", "high": "🟠", "medium": "🟡", "low": "🟢"}[sev]
			sb.WriteString(fmt.Sprintf("%s %s (%d):\n", icon, strings.ToUpper(sev), len(vulns)))
			for _, v := range vulns {
				sb.WriteString(fmt.Sprintf("• %s - %s@%s\n", v.ID, v.Package, v.Version))
				sb.WriteString("[hr]\n")
			}
		}
	}

	sb.WriteString("[/info]")
	return sb.String()
}

type OSVQuery struct {
	Package OSPackage `json:"package"`
	Version string    `json:"version"`
}

type OSPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type OSVBatchRequest struct {
	Queries []OSVQuery `json:"queries"`
}

type OSVResponse struct {
	Results []OSVResult `json:"results"`
}

type OSVResult struct {
	Vulns []OSVVuln `json:"vulns"`
}

type OSVVuln struct {
	ID               string         `json:"id"`
	Summary          string         `json:"summary"`
	Severity         any            `json:"severity"`
	FixedVersion     string         `json:"fixed_version"`
	Affected         []OSVAffected  `json:"affected"`
	Aliases          []string       `json:"aliases"`
	References       []OSVReference `json:"references"`
	DatabaseSpecific any            `json:"database_specific"`
}

type OSVDetail struct {
	Type     string  `json:"type"`
	Severity string  `json:"severity"`
	Score    float64 `json:"score"`
}

type OSVReference struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type OSVAffected struct {
	Package OSPackage   `json:"package"`
	Ranges  []OSVRanges `json:"ranges"`
}

type OSVRanges struct {
	Type   string      `json:"type"`
	Events []OSVEvents `json:"events"`
}

type OSVEvents struct {
	Introduced string `json:"introduced"`
	Fixed      string `json:"fixed"`
}

func generateUUID() string {
	return fmt.Sprintf("cve-%s", utils.GenerateRandomString(16))
}

func getVulnDetails(id string) OSVVuln {
	url := fmt.Sprintf("https://api.osv.dev/v1/vulns/%s", id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Warnf("OSV: failed to create request for %s: %v", id, err)
		return OSVVuln{}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warnf("OSV: failed to fetch %s: %v", id, err)
		return OSVVuln{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("OSV: status %d for %s", resp.StatusCode, id)
		return OSVVuln{}
	}

	var vuln OSVVuln
	if err := json.NewDecoder(resp.Body).Decode(&vuln); err != nil {
		logger.Warnf("OSV: failed to decode %s: %v", id, err)
		return OSVVuln{}
	}

	summary := vuln.Summary
	if len(summary) > 100 {
		summary = summary[:100]
	}

	return vuln
}

func extractScore(v OSVVuln) float64 {
	// Try database_specific first - severity maps to score reliably
	if v.DatabaseSpecific != nil {
		switch ds := v.DatabaseSpecific.(type) {
		case map[string]any:
			if severity, ok := ds["severity"].(string); ok {
				return mapSeverityToScore(severity)
			}
		}
	}

	// Fallback to CVSS vector parsing from severity array
	switch s := v.Severity.(type) {
	case []any:
		for _, item := range s {
			if sevMap, ok := item.(map[string]any); ok {
				if scoreStr, ok := sevMap["score"].(string); ok {
					return parseCVSSVector(scoreStr)
				}
			}
		}
	case map[string]any:
		if scoreStr, ok := s["score"].(string); ok {
			return parseCVSSVector(scoreStr)
		}
	}

	return 0
}

func parseCVSSVector(cvssVector string) float64 {
	// Parse CVSS 3.1 vector string to calculate base score
	// Vector format: "CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:N/A:N"

	// Default impact scores
	impactConfidential := 0.0
	impactIntegrity := 0.0
	impactAvailability := 0.0

	// Parse impact metrics (I:C, I:I, I:A)
	impactValues := map[string]float64{"H": 0.56, "M": 0.22, "L": 0.06}
	for imp, val := range impactValues {
		if strings.Contains(cvssVector, "/C:"+imp) {
			impactConfidential = val
		}
		if strings.Contains(cvssVector, "/I:"+imp) {
			impactIntegrity = val
		}
		if strings.Contains(cvssVector, "/A:"+imp) {
			impactAvailability = val
		}
	}

	// Attack complexity
	acValue := 0.44
	if strings.Contains(cvssVector, "/AC:L") {
		acValue = 0.77
	}

	// Attack vector
	avValue := 0.85
	if strings.Contains(cvssVector, "/AV:N") {
		avValue = 0.55
	} else if strings.Contains(cvssVector, "/AV:A") {
		avValue = 0.62
	} else if strings.Contains(cvssVector, "/AV:L") {
		avValue = 0.2
	}

	// Privileges required
	prValue := 0.85
	if strings.Contains(cvssVector, "/PR:N") {
		prValue = 0.85 // N/A in CVSS 3.1 for network is 0.85
	} else if strings.Contains(cvssVector, "/PR:L") {
		prValue = 0.62
	} else if strings.Contains(cvssVector, "/PR:H") {
		prValue = 0.27
	}

	// User interaction
	uiValue := 0.85
	if strings.Contains(cvssVector, "/UI:R") {
		uiValue = 0.62
	}

	// Scope change check
	scopeChanged := strings.Contains(cvssVector, "S:C")

	// Simplified CVSS 3.1 calculation
	var impact float64
	if scopeChanged {
		impact = 1 - ((1 - impactConfidential) * 0.25 * (1 - impactIntegrity) * 0.25 * (1 - impactAvailability) * 0.25)
	} else {
		impact = 1 - (1-impactConfidential)*(1-impactIntegrity)*(1-impactAvailability)
	}

	exploitability := 8.22 * avValue * acValue * prValue * uiValue
	var baseScore float64
	if impact <= 0 {
		baseScore = 0
	} else {
		if scopeChanged {
			baseScore = 1.08 * (impact + exploitability)
		} else {
			baseScore = impact + exploitability
		}
	}

	if baseScore > 10 {
		baseScore = 10
	}

	// Round to nearest 0.1
	baseScore = float64(int(baseScore*10+0.5)) / 10

	// Apply severity thresholds
	if baseScore >= 9.0 {
		return 9.3
	} else if baseScore >= 7.0 {
		return 8.3
	} else if baseScore >= 4.0 {
		return 5.3
	} else if baseScore >= 0.1 {
		return 2.8
	}

	return baseScore
}

func extractSeverity(v OSVVuln) string {
	// Try database_specific first - most reliable source
	if v.DatabaseSpecific != nil {
		switch ds := v.DatabaseSpecific.(type) {
		case map[string]any:
			if sev, ok := ds["severity"].(string); ok && sev != "" {
				return strings.ToUpper(sev)
			}
		}
	}

	// Try CVSS score as fallback
	cvssScore := extractCVSSScoreFromVuln(v)
	if cvssScore > 0 {
		return mapScoreToSeverity(cvssScore)
	}

	// Try from severity array (CVSS type)
	switch s := v.Severity.(type) {
	case []any:
		for _, item := range s {
			if sevMap, ok := item.(map[string]any); ok {
				if sev, ok := sevMap["type"].(string); ok && sev != "" {
					if strings.Contains(sev, "CVSS") {
						continue
					}
					return sev
				}
			}
		}
	case map[string]any:
		if sev, ok := s["type"].(string); ok && sev != "" {
			return sev
		}
	}

	return ""
}

func extractCVSSScoreFromVuln(v OSVVuln) float64 {
	switch s := v.Severity.(type) {
	case []any:
		for _, item := range s {
			if sevMap, ok := item.(map[string]any); ok {
				if scoreStr, ok := sevMap["score"].(string); ok && scoreStr != "" {
					return parseCVSSVector(scoreStr)
				}
			}
		}
	case map[string]any:
		if scoreStr, ok := s["score"].(string); ok && scoreStr != "" {
			return parseCVSSVector(scoreStr)
		}
	}
	return 0
}

func mapScoreToSeverity(score float64) string {
	if score >= 9.0 {
		return "CRITICAL"
	} else if score >= 7.0 {
		return "HIGH"
	} else if score >= 4.0 {
		return "MEDIUM"
	} else if score >= 0.1 {
		return "LOW"
	}
	return "UNKNOWN"
}

func mapSeverityToScore(severity string) float64 {
	switch strings.ToUpper(severity) {
	case "CRITICAL":
		return 9.0
	case "HIGH":
		return 7.0
	case "MODERATE":
		return 5.0
	case "LOW":
		return 3.0
	default:
		return 0
	}
}
