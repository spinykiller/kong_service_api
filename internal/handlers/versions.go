package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yashjain/konnect/internal/database"
	"github.com/yashjain/konnect/internal/models"
	"github.com/yashjain/konnect/pkg/types"
	"github.com/yashjain/konnect/pkg/utils"
)

// GetVersions godoc
// @Summary Get versions for a service
// @Description Get a paginated list of versions for a specific service
// @Tags versions
// @Produce json
// @Param id path string true "Service ID"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} types.PaginatedResponse{data=[]models.Version}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id}/versions [get]
func GetVersions(c *gin.Context) {
	serviceID := c.Param("id")

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

	// Get versions from database
	versions, total, err := database.GetVersions(serviceID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create paginated response
	pagination := utils.CalculatePagination(params.Page, params.PageSize, total)
	response := types.PaginatedResponse{
		Data:       versions,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// CreateVersion godoc
// @Summary Create a new version
// @Description Create a new version for a specific service
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param version body models.Version true "Version object"
// @Success 201 {object} models.Version
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id}/versions [post]
func CreateVersion(c *gin.Context) {
	serviceID := c.Param("id")

	var version models.Version
	if err := c.ShouldBindJSON(&version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	version.ID = uuid.New().String()
	version.ServiceID = serviceID

	err := database.CreateVersion(&version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}
