package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type CveConfigHandler struct {
	service     services.ICveConfigService
	cronService services.ICronService
}

func NewCveConfigHandler(service services.ICveConfigService, cronService services.ICronService) *CveConfigHandler {
	return &CveConfigHandler{
		service:     service,
		cronService: cronService,
	}
}

func (h *CveConfigHandler) GetByProject(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	paging := utils.GeneratePagingFromRequest(c)

	configs, total, err := h.service.GetByProjectID(uint(projectID), paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	data := make([]gin.H, 0, len(configs))
	for i := range configs {
		data = append(data, buildCveConfigResponse(&configs[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

func (h *CveConfigHandler) Create(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	var input struct {
		Name            string `json:"name" binding:"required"`
		RepoUrl         string `json:"repoUrl"`
		Languages       string `json:"languages" binding:"required"`
		Cron            string `json:"cron" binding:"required"`
		Status          string `json:"status"`
		ApiKey          string `json:"apiKey"`
		BotID           *int   `json:"botId"`
		NotifyOnSuccess *bool  `json:"notifyOnSuccess"`
		NotifyOnFailure *bool  `json:"notifyOnFailure"`
		NotifyRoomId    string `json:"notifyRoomId"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	notifyOnSuccess := false
	notifyOnFailure := true
	if input.NotifyOnSuccess != nil {
		notifyOnSuccess = *input.NotifyOnSuccess
	}
	if input.NotifyOnFailure != nil {
		notifyOnFailure = *input.NotifyOnFailure
	}
	serviceInput := &services.CveConfigInput{
		Name:            input.Name,
		RepoUrl:         input.RepoUrl,
		Languages:       input.Languages,
		Cron:            input.Cron,
		Status:          input.Status,
		ApiKey:          input.ApiKey,
		BotID:           input.BotID,
		NotifyOnSuccess: notifyOnSuccess,
		NotifyOnFailure: notifyOnFailure,
		NotifyRoomId:    input.NotifyRoomId,
	}

	config, err := h.service.Create(uint(projectID), serviceInput)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseInsert, err.Error()))
		return
	}

	h.cronService.SyncCVEConfigs()

	utils.RespondWithOK(c, http.StatusCreated, buildCveConfigResponse(config))
}

func (h *CveConfigHandler) Update(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	var input struct {
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

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	serviceInput := &services.CveConfigUpdateInput{
		Name:            input.Name,
		RepoUrl:         input.RepoUrl,
		Languages:       input.Languages,
		Cron:            input.Cron,
		Status:          input.Status,
		ApiKey:          input.ApiKey,
		BotID:           input.BotID,
		NotifyOnSuccess: input.NotifyOnSuccess,
		NotifyOnFailure: input.NotifyOnFailure,
		NotifyRoomId:    input.NotifyRoomId,
	}

	config, err := h.service.Update(configID, uint(projectID), serviceInput)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseUpdate, err.Error()))
		return
	}

	h.cronService.SyncCVEConfigs()

	utils.RespondWithOK(c, http.StatusOK, buildCveConfigResponse(config))
}

func (h *CveConfigHandler) Delete(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	if err := h.service.Delete(configID, uint(projectID)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseDelete, err.Error()))
		return
	}

	h.cronService.SyncCVEConfigs()

	c.Status(http.StatusNoContent)
}

func (h *CveConfigHandler) Toggle(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	config, err := h.service.Toggle(configID, uint(projectID))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Config not found"))
		return
	}

	h.cronService.SyncCVEConfigs()

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"id":     config.ID,
		"status": config.Status,
	})
}

func (h *CveConfigHandler) Scan(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	if err := h.service.TriggerScan(configID, uint(projectID)); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(2002, err.Error()))
		return
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"message": "Scan triggered successfully",
	})
}

func (h *CveConfigHandler) Test(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	var input struct {
		Languages string `json:"languages" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	vulns, err := h.service.TestScan(input.Languages)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(2002, err.Error()))
		return
	}

	data := make([]gin.H, 0, len(vulns))
	for i := range vulns {
		data = append(data, buildVulnerabilityResponse(&vulns[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": len(vulns),
	})
}

func (h *CveConfigHandler) TestPublic(c *gin.Context) {
	var input struct {
		Languages string `json:"languages" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))
		return
	}

	vulns, err := h.service.TestScan(input.Languages)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(2002, err.Error()))
		return
	}

	data := make([]gin.H, 0, len(vulns))
	for i := range vulns {
		data = append(data, buildVulnerabilityResponse(&vulns[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": len(vulns),
	})
}

func (h *CveConfigHandler) GetVulnerabilities(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	vulns, total, err := h.service.GetVulnerabilities(configID, uint(projectID))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Config not found"))
		return
	}

	data := make([]gin.H, 0, len(vulns))
	for i := range vulns {
		data = append(data, buildVulnerabilityResponse(&vulns[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": total,
	})
}

func (h *CveConfigHandler) GetScanLogs(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	configID := c.Param("configId")
	if configID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, "configId is required"))
		return
	}

	paging := utils.GeneratePagingFromRequest(c)

	logs, total, err := h.service.GetScanLogs(configID, uint(projectID), paging)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, errors.New(errors.ErrResourceNotFound, "Config not found"))
		return
	}

	data := make([]gin.H, 0, len(logs))
	for i := range logs {
		data = append(data, buildCveScanLogResponse(&logs[i]))
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  paging.Page,
		"limit": paging.Limit,
	})
}

func (h *CveConfigHandler) GetAnalysis(c *gin.Context) {
	projectID, err := parseIDParam(c, "projectId")
	if err != nil {
		return
	}

	if err := h.checkProjectAccess(c, uint(projectID)); err != nil {
		return
	}

	analysis, err := h.service.GetAnalysisByProject(uint(projectID))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, errors.New(errors.ErrDatabaseQuery, err.Error()))
		return
	}

	data := make([]gin.H, 0, len(analysis))
	for i := range analysis {
		vulns := make([]gin.H, 0, len(analysis[i].Vulnerabilities))
		for j := range analysis[i].Vulnerabilities {
			vulns = append(vulns, buildVulnerabilityResponse(&analysis[i].Vulnerabilities[j]))
		}

		item := gin.H{
			"configId":        analysis[i].ConfigID,
			"configName":      analysis[i].ConfigName,
			"configStatus":    analysis[i].ConfigStatus,
			"lastScan":        analysis[i].LastScan,
			"lastStatus":      analysis[i].LastStatus,
			"vulnerabilities": vulns,
		}
		data = append(data, item)
	}

	utils.RespondWithOK(c, http.StatusOK, gin.H{
		"data":  data,
		"total": len(data),
	})
}

func (h *CveConfigHandler) checkProjectAccess(c *gin.Context, projectID uint) error {
	projectSecretKey := c.GetHeader("X-Project-Key")
	if projectSecretKey == "" {
		projectSecretKey = c.Query("project_key")
	}

	if projectSecretKey != "" {
		return nil
	}

	_, exists := c.Get("projectID")
	if exists {
		return nil
	}

	utils.RespondWithError(c, http.StatusUnauthorized, errors.New(errors.ErrAuthUnauthorized, "Unauthorized"))
	return &errors.AppError{Code: errors.ErrAuthUnauthorized, Message: "Unauthorized"}
}

func buildCveScanLogResponse(log *models.CveScanLog) gin.H {
	resp := gin.H{
		"id":             log.ID,
		"configId":       log.ConfigID,
		"projectId":      log.ProjectID,
		"status":         log.Status,
		"vulnFoundCount": log.VulnFoundCount,
		"startedAt":      log.StartedAt.Format("2006-01-02T15:04:05Z"),
	}

	if log.FinishedAt != nil {
		resp["finishedAt"] = log.FinishedAt.Format("2006-01-02T15:04:05Z")
	}

	if log.ErrorMessage != "" {
		resp["errorMessage"] = log.ErrorMessage
	}

	return resp
}

func buildCveConfigResponse(config *models.CveConfig) gin.H {
	resp := gin.H{
		"id":                   config.ID,
		"projectId":            config.ProjectID,
		"name":                 config.Name,
		"repoUrl":              config.RepoUrl,
		"languages":            config.Languages,
		"cron":                 config.Cron,
		"status":               config.Status,
		"lastScan":             config.LastScan,
		"lastStatus":           config.LastStatus,
		"vulnerabilitiesFound": config.VulnerabilitiesFound,
		"createdAt":            config.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if config.BotID != nil {
		resp["botId"] = *config.BotID
	}
	resp["notifyOnSuccess"] = config.NotifyOnSuccess
	resp["notifyOnFailure"] = config.NotifyOnFailure
	if config.NotifyRoomId != "" {
		resp["notifyRoomId"] = config.NotifyRoomId
	}
	if config.ApiKey != "" {
		resp["apiKey"] = "cwk_***hidden***"
	}

	return resp
}

func buildVulnerabilityResponse(vuln *models.Vulnerability) gin.H {
	resp := gin.H{
		"id":        vuln.ID,
		"scanLogId": vuln.ScanLogID,
		"configId":  vuln.ConfigID,
		"cveId":     vuln.CVEID,
		"severity":  vuln.Severity,
		"package":   vuln.Package,
		"version":   vuln.Version,
		"summary":   vuln.Summary,
	}

	if vuln.Score > 0 {
		resp["score"] = vuln.Score
	}

	if vuln.ReferenceURL != "" {
		resp["referenceUrl"] = vuln.ReferenceURL
	}

	return resp
}
