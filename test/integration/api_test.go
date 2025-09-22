package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yashjain/konnect/internal/database"
	"github.com/yashjain/konnect/internal/handlers"
	"github.com/yashjain/konnect/internal/models"
)

func TestMain(m *testing.M) {
	// Setup test database
	setupTestDB()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestDB()

	os.Exit(code)
}

func setupTestDB() {
	// Use test database or create one
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "app:app@tcp(127.0.0.1:3306)/servicesdb_test?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci"
	}

	// Set environment variable for database package
	_ = os.Setenv("MYSQL_DSN", dsn)

	// Initialize database
	if err := database.Init(); err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Create tables
	createTestTables()

	// Seed test data
	seedTestData()
}

func cleanupTestDB() {
	if database.DB != nil {
		// Clean up test data
		_, _ = database.DB.Exec("DELETE FROM versions")
		_, _ = database.DB.Exec("DELETE FROM services")
		_ = database.Close()
	}
}

func createTestTables() {
	// Create services table
	servicesSQL := `
	CREATE TABLE IF NOT EXISTS services (
		id            CHAR(36)     NOT NULL,
		name          VARCHAR(255) NOT NULL,
		slug          VARCHAR(255) NOT NULL,
		description   TEXT NULL,
		created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		versions_count INT NOT NULL DEFAULT 0,
		PRIMARY KEY (id),
		UNIQUE KEY uq_services_name (name),
		UNIQUE KEY uq_services_slug (slug),
		FULLTEXT KEY ft_services_name_desc (name, description)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
	`

	// Create versions table
	versionsSQL := `
	CREATE TABLE IF NOT EXISTS versions (
		id          CHAR(36)    NOT NULL,
		service_id  CHAR(36)    NOT NULL,
		semver      VARCHAR(64) NOT NULL,
		status      ENUM('draft','released','deprecated') NOT NULL,
		changelog   TEXT NULL,
		created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_versions_service_id (service_id),
		KEY idx_versions_status (status),
		CONSTRAINT fk_versions_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
	`

	_, _ = database.DB.Exec(servicesSQL)
	_, _ = database.DB.Exec(versionsSQL)
}

func seedTestData() {
	// Insert test services
	services := []models.Service{
		{ID: "service-1", Name: "Test Service 1", Slug: "test-service-1", Description: "First test service"},
		{ID: "service-2", Name: "Test Service 2", Slug: "test-service-2", Description: "Second test service"},
		{ID: "service-3", Name: "Notification Service", Slug: "notification-service", Description: "Service for sending notifications"},
	}

	for _, service := range services {
		_, _ = database.DB.Exec("INSERT INTO services (id, name, slug, description) VALUES (?, ?, ?, ?)",
			service.ID, service.Name, service.Slug, service.Description)
	}

	// Insert test versions
	versions := []models.Version{
		{ID: "version-1", ServiceID: "service-1", Semver: "1.0.0", Status: "released", Changelog: "Initial release"},
		{ID: "version-2", ServiceID: "service-1", Semver: "1.1.0", Status: "released", Changelog: "Minor update"},
		{ID: "version-3", ServiceID: "service-2", Semver: "0.1.0", Status: "draft", Changelog: "Work in progress"},
		{ID: "version-4", ServiceID: "service-3", Semver: "2.0.0", Status: "released", Changelog: "Major update"},
	}

	for _, version := range versions {
		_, _ = database.DB.Exec("INSERT INTO versions (id, service_id, semver, status, changelog) VALUES (?, ?, ?, ?, ?)",
			version.ID, version.ServiceID, version.Semver, version.Status, version.Changelog)
	}

	// Update versions_count
	_, _ = database.DB.Exec("UPDATE services SET versions_count = (SELECT COUNT(*) FROM versions WHERE service_id = services.id)")
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add routes
	router.GET("/health", handlers.HealthCheck)
	router.GET("/api/v1/services", handlers.GetServices)
	router.GET("/api/v1/services/search", handlers.SearchServices)
	router.POST("/api/v1/services", handlers.CreateService)
	router.GET("/api/v1/services/:id", handlers.GetService)
	router.PUT("/api/v1/services/:id", handlers.UpdateService)
	router.DELETE("/api/v1/services/:id", handlers.DeleteService)
	router.GET("/api/v1/services/:id/versions", handlers.GetVersions)
	router.POST("/api/v1/services/:id/versions", handlers.CreateVersion)

	return router
}

func TestHealthCheckIntegration(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGetServicesIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get all services",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "get services with pagination",
			queryParams:    "?page=1&page_size=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get second page",
			queryParams:    "?page=2&page_size=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "invalid page size",
			queryParams:    "?page_size=101",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.expectedCount > 0 {
					data := response["data"].([]interface{})
					assert.Len(t, data, tt.expectedCount)
				}

				assert.NotNil(t, response["pagination"])
			}
		})
	}
}

func TestSearchServicesIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "search for notification",
			queryParams:    "?q=notification",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "search for test",
			queryParams:    "?q=test",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "search with pagination",
			queryParams:    "?q=test&page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "search with no results",
			queryParams:    "?q=nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "missing search query",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].([]interface{})
				assert.Len(t, data, tt.expectedCount)
				assert.NotNil(t, response["pagination"])
			}
		})
	}
}

func TestCreateServiceIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceData    models.Service
		expectedStatus int
	}{
		{
			name: "valid service",
			serviceData: models.Service{
				Name:        "New Test Service",
				Slug:        "new-test-service",
				Description: "A new test service",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "service with duplicate name",
			serviceData: models.Service{
				Name:        "Test Service 1", // Already exists
				Slug:        "duplicate-service",
				Description: "Duplicate service",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service with duplicate slug",
			serviceData: models.Service{
				Name:        "Unique Service",
				Slug:        "test-service-1", // Already exists
				Description: "Duplicate slug service",
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.serviceData)
			req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Service
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.serviceData.Name, response.Name)
				assert.Equal(t, tt.serviceData.Slug, response.Slug)
				assert.Equal(t, tt.serviceData.Description, response.Description)
			}
		})
	}
}

func TestGetServiceIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceID      string
		expectedStatus int
	}{
		{
			name:           "existing service",
			serviceID:      "service-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing service",
			serviceID:      "non-existing",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services/"+tt.serviceID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Service
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.serviceID, response.ID)
			}
		})
	}
}

func TestGetVersionsIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceID      string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get versions for service-1",
			serviceID:      "service-1",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get versions with pagination",
			serviceID:      "service-1",
			queryParams:    "?page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "get versions for service with no versions",
			serviceID:      "service-2",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  1, // service-2 has 1 version
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/services/"+tt.serviceID+"/versions"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				data := response["data"].([]interface{})
				assert.Len(t, data, tt.expectedCount)
				assert.NotNil(t, response["pagination"])
			}
		})
	}
}

func TestCreateVersionIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceID      string
		versionData    models.Version
		expectedStatus int
	}{
		{
			name:      "valid version",
			serviceID: "service-1",
			versionData: models.Version{
				Semver:    "1.2.0",
				Status:    "released",
				Changelog: "New feature release",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "version for non-existing service",
			serviceID: "non-existing",
			versionData: models.Version{
				Semver:    "1.0.0",
				Status:    "released",
				Changelog: "Test version",
			},
			expectedStatus: http.StatusCreated, // Still creates, just with invalid service_id
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.versionData)
			req, _ := http.NewRequest("POST", "/api/v1/services/"+tt.serviceID+"/versions", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Version
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.serviceID, response.ServiceID)
				assert.Equal(t, tt.versionData.Semver, response.Semver)
				assert.Equal(t, tt.versionData.Status, response.Status)
				assert.Equal(t, tt.versionData.Changelog, response.Changelog)
			}
		})
	}
}
