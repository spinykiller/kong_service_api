package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/yashjain/konnect/docs"
)

type Service struct {
	ID            string `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Slug          string `json:"slug" db:"slug"`
	Description   string `json:"description" db:"description"`
	CreatedAt     string `json:"created_at" db:"created_at"`
	UpdatedAt     string `json:"updated_at" db:"updated_at"`
	VersionsCount int    `json:"versions_count" db:"versions_count"`
}

type Version struct {
	ID        string `json:"id" db:"id"`
	ServiceID string `json:"service_id" db:"service_id"`
	Semver    string `json:"semver" db:"semver"`
	Status    string `json:"status" db:"status"`
	Changelog string `json:"changelog" db:"changelog"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

// Pagination structures
type PaginationParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// Search parameters
type SearchParams struct {
	Query    string `form:"q" binding:"required"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

var db *sql.DB

// Helper function to get pagination parameters with defaults
func getPaginationParams(c *gin.Context) PaginationParams {
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Parse page_size parameter
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	return params
}

// Helper function to get search parameters with defaults
func getSearchParams(c *gin.Context) SearchParams {
	params := SearchParams{
		Query:    c.Query("q"),
		Page:     1,
		PageSize: 10,
	}

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Parse page_size parameter
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	return params
}

// Helper function to calculate pagination metadata
func calculatePagination(page, pageSize, total int) Pagination {
	totalPages := (total + pageSize - 1) / pageSize // Ceiling division

	return Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// @title Services API
// @version 1.0
// @description A REST API for managing services and their versions
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	// Initialize database connection
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci"
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
		log.Fatal("Failed to ping database:", err)
	}

	// Set up Gin router
	if os.Getenv("LOG_LEVEL") == "info" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	r.GET("/health", healthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/services", getServices)
		api.GET("/services/search", searchServices)
		api.POST("/services", createService)
		api.GET("/services/:id", getService)
		api.PUT("/services/:id", updateService)
		api.DELETE("/services/:id", deleteService)

		api.GET("/services/:id/versions", getVersions)
		api.POST("/services/:id/versions", createVersion)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up database cleanup before starting server
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}

// healthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// getServices godoc
// @Summary Get all services
// @Description Get a paginated list of all services
// @Tags services
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} PaginatedResponse{data=[]Service}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services [get]
func getServices(c *gin.Context) {
	// Get pagination parameters
	params := getPaginationParams(c)

	// Validate pagination parameters
	if params.Page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be greater than 0"})
		return
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be between 1 and 100"})
		return
	}

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM services").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated services
	query := "SELECT id, name, slug, description, created_at, updated_at, versions_count FROM services ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := db.Query(query, params.PageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var services []Service
	for rows.Next() {
		var s Service
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.VersionsCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		services = append(services, s)
	}

	// Create paginated response
	pagination := calculatePagination(params.Page, params.PageSize, total)
	response := PaginatedResponse{
		Data:       services,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// searchServices godoc
// @Summary Search services
// @Description Search services by name, slug, or description using full-text search
// @Tags services
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} PaginatedResponse{data=[]Service}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/search [get]
func searchServices(c *gin.Context) {
	// Get search parameters
	params := getSearchParams(c)

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

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Get total count for search results
	countQuery := "SELECT COUNT(*) FROM services WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)"
	var total int
	err := db.QueryRow(countQuery, params.Query).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated search results
	searchQuery := `
		SELECT id, name, slug, description, created_at, updated_at, versions_count 
		FROM services 
		WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)
		ORDER BY MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE) DESC, created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(searchQuery, params.Query, params.Query, params.PageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var services []Service
	for rows.Next() {
		var s Service
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.VersionsCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		services = append(services, s)
	}

	// Create paginated response
	pagination := calculatePagination(params.Page, params.PageSize, total)
	response := PaginatedResponse{
		Data:       services,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// createService godoc
// @Summary Create a new service
// @Description Create a new service with the provided information
// @Tags services
// @Accept json
// @Produce json
// @Param service body Service true "Service object"
// @Success 201 {object} Service
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services [post]
func createService(c *gin.Context) {
	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service.ID = uuid.New().String()

	_, err := db.Exec("INSERT INTO services (id, name, slug, description) VALUES (?, ?, ?, ?)",
		service.ID, service.Name, service.Slug, service.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// getService godoc
// @Summary Get a service by ID
// @Description Get a specific service by its ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} Service
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [get]
func getService(c *gin.Context) {
	id := c.Param("id")

	var service Service
	err := db.QueryRow("SELECT id, name, slug, description, created_at, updated_at, versions_count FROM services WHERE id = ?", id).
		Scan(&service.ID, &service.Name, &service.Slug, &service.Description, &service.CreatedAt, &service.UpdatedAt, &service.VersionsCount)

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

// updateService godoc
// @Summary Update a service
// @Description Update a service with the provided information
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param service body Service true "Service object"
// @Success 200 {object} Service
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [put]
func updateService(c *gin.Context) {
	id := c.Param("id")

	var service Service
	if err := c.ShouldBindJSON(&service); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("UPDATE services SET name = ?, slug = ?, description = ? WHERE id = ?",
		service.Name, service.Slug, service.Description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	service.ID = id
	c.JSON(http.StatusOK, service)
}

// deleteService godoc
// @Summary Delete a service
// @Description Delete a service by its ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id} [delete]
func deleteService(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM services WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted"})
}

// getVersions godoc
// @Summary Get versions for a service
// @Description Get a paginated list of versions for a specific service
// @Tags versions
// @Produce json
// @Param id path string true "Service ID"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Number of items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} PaginatedResponse{data=[]Version}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id}/versions [get]
func getVersions(c *gin.Context) {
	serviceID := c.Param("id")

	// Get pagination parameters
	params := getPaginationParams(c)

	// Validate pagination parameters
	if params.Page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be greater than 0"})
		return
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size must be between 1 and 100"})
		return
	}

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Get total count for this service
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM versions WHERE service_id = ?", serviceID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated versions
	query := "SELECT id, service_id, semver, status, changelog, created_at FROM versions WHERE service_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := db.Query(query, serviceID, params.PageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var versions []Version
	for rows.Next() {
		var v Version
		err := rows.Scan(&v.ID, &v.ServiceID, &v.Semver, &v.Status, &v.Changelog, &v.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		versions = append(versions, v)
	}

	// Create paginated response
	pagination := calculatePagination(params.Page, params.PageSize, total)
	response := PaginatedResponse{
		Data:       versions,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// createVersion godoc
// @Summary Create a new version
// @Description Create a new version for a specific service
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param version body Version true "Version object"
// @Success 201 {object} Version
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /services/{id}/versions [post]
func createVersion(c *gin.Context) {
	serviceID := c.Param("id")

	var version Version
	if err := c.ShouldBindJSON(&version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	version.ID = uuid.New().String()
	version.ServiceID = serviceID

	// Start a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("Error rolling back transaction: %v", err)
		}
	}()

	// Insert the version
	_, err = tx.Exec("INSERT INTO versions (id, service_id, semver, status, changelog) VALUES (?, ?, ?, ?, ?)",
		version.ID, version.ServiceID, version.Semver, version.Status, version.Changelog)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the versions_count in the services table
	_, err = tx.Exec("UPDATE services SET versions_count = versions_count + 1 WHERE id = ?", serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}
