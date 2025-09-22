package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yashjain/konnect/internal/database"
	"github.com/yashjain/konnect/internal/models"
	"github.com/yashjain/konnect/pkg/types"
	"github.com/yashjain/konnect/pkg/utils"
)

// GetServices godoc
// @Summary Get all services
// @Description Get a paginated list of all services
// @Tags services
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} types.PaginatedResponse{data=[]models.Service}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services [get]
func GetServices(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Validate pagination parameters
	if params.Page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be greater than 0"})
		return
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be between 1 and 100"})
		return
	}

	// Get services from database
	services, total, err := database.GetServices(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create paginated response
	pagination := utils.CalculatePagination(params.Page, params.PageSize, total)
	response := types.PaginatedResponse{
		Data:       services,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// SearchServices godoc
// @Summary Search services
// @Description Search services by name, slug, or description using full-text search
// @Tags services
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} types.PaginatedResponse{data=[]models.Service}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/search [get]
func SearchServices(c *gin.Context) {
	// Get search parameters
	params := utils.GetSearchParams(c)

	// Validate search query
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	// Validate pagination parameters
	if params.Page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be greater than 0"})
		return
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be between 1 and 100"})
		return
	}

	// Search services in database
	services, total, err := database.SearchServices(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create paginated response
	pagination := utils.CalculatePagination(params.Page, params.PageSize, total)
	response := types.PaginatedResponse{
		Data:       services,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// CreateService godoc
// @Summary Create a new service
// @Description Create a new service with the provided information
// @Tags services
// @Accept json
// @Produce json
// @Param service body models.Service true "Service object"
// @Success 201 {object} models.Service
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services [post]
func CreateService(c *gin.Context) {
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.ID = uuid.New().String()

	err := database.CreateService(&service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// GetService godoc
// @Summary Get a service by ID
// @Description Get a specific service by its ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} models.Service
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [get]
func GetService(c *gin.Context) {
	id := c.Param("id")

	service, err := database.GetServiceByID(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service)
}

// UpdateService godoc
// @Summary Update a service
// @Description Update a service with the provided information
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param service body models.Service true "Service object"
// @Success 200 {object} models.Service
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [put]
func UpdateService(c *gin.Context) {
	id := c.Param("id")

	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := database.UpdateService(id, &service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	service.ID = id
	c.JSON(http.StatusOK, service)
}

// DeleteService godoc
// @Summary Delete a service
// @Description Delete a service by its ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [delete]
func DeleteService(c *gin.Context) {
	id := c.Param("id")

	rowsAffected, err := database.DeleteService(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted"})
}
