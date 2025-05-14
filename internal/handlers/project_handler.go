package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

type IProjectHandler interface {
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	VerifyAccess(c *gin.Context)
}

type ProjectHandler struct {
	service     services.IProjectService
	cronService services.ICronService
}

func NewProjectHandler(
	service services.IProjectService,
	cronService services.ICronService,
) *ProjectHandler {
	return &ProjectHandler{
		service:     service,
		cronService: cronService,
	}
}

func (h *ProjectHandler) GetAll(c *gin.Context) {
	projects, err := h.service.GetAll()
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadGateway,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
	}
	utils.RespondWithOK(c, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"required,max=255"`
		SecretKey   string `json:"secret_key"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, err.Error()),
		)
		return
	}

	project := &models.Project{
		Name:        input.Name,
		Description: input.Description,
		SecretKey:   input.SecretKey,
	}
	createdProject, err := h.service.Create(project)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadGateway,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	utils.RespondWithOK(c, http.StatusCreated, createdProject)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	projectId := c.Param("id")

	// Check if project ID is valid
	id, err := strconv.Atoi(projectId)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	res := utils.CensorSensitiveData(project, []string{"secretKey"})

	utils.RespondWithOK(c, http.StatusOK, res)
}

func (h *ProjectHandler) Update(c *gin.Context) {

	// Get role ID from URL parameter
	projectId := c.Param("id")

	// Check if project ID is valid
	id, err := strconv.Atoi(projectId)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	var input struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description" binding:"required,max=255"`
		SecretKey   *string `json:"secret_key"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, err.Error()),
		)
		return
	}

	// Update project fields
	project.Name = input.Name
	project.Description = input.Description

	if input.SecretKey != nil {
		project.SecretKey = *input.SecretKey
	}

	updatedProject, err := h.service.Update(project)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadGateway,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}
	utils.RespondWithOK(c, http.StatusOK, updatedProject)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	// Get role ID from URL parameter
	projectId := c.Param("id")

	// Check if project ID is valid
	id, err := strconv.Atoi(projectId)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidParse, err.Error()),
		)
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil || project == nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	err = h.service.Delete(uint(id))
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadGateway,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	h.cronService.SyncAll()

	utils.RespondWithOK(c, http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *ProjectHandler) VerifyAccess(c *gin.Context) {
	var input struct {
		ProjectID uint   `json:"project_id" binding:"required"`
		SecretKey string `json:"secret_key" binding:"required,max=255"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrInvalidData, err.Error()),
		)
		return
	}
	project, err := h.service.GetByID(input.ProjectID)
	if err != nil {
		utils.RespondWithError(
			c,
			http.StatusBadRequest,
			errors.New(errors.ErrDatabaseQuery, err.Error()),
		)
		return
	}

	if project.SecretKey != input.SecretKey {
		utils.RespondWithOK(
			c,
			http.StatusOK, gin.H{
				"isAccess": false,
				"message":  "Access denied",
			})
	} else {
		utils.RespondWithOK(
			c,
			http.StatusOK, gin.H{
				"isAccess": true,
				"message":  "Access granted",
			})
	}

}
