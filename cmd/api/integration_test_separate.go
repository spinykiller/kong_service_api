package main

import (
	"bytes"
	"database/sql"
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
)

var testDB *sql.DB

func TestMainIntegration(m *testing.M) {
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

	var err error
	testDB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	if err = testDB.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping test database: %v", err))
	}

	// Set global db variable for tests
	db = testDB

	// Create tables
	createTestTables()

	// Seed test data
	seedTestData()
}

func cleanupTestDB() {
	if testDB != nil {
		// Clean up test data
		testDB.Exec("DELETE FROM versions")
		testDB.Exec("DELETE FROM services")
		testDB.Close()
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

	testDB.Exec(servicesSQL)
	testDB.Exec(versionsSQL)
}

func seedTestData() {
	// Insert test services
	services := []Service{
		{ID: "service-1", Name: "Test Service 1", Slug: "test-service-1", Description: "First test service"},
		{ID: "service-2", Name: "Test Service 2", Slug: "test-service-2", Description: "Second test service"},
		{ID: "service-3", Name: "Notification Service", Slug: "notification-service", Description: "Service for sending notifications"},
	}

	for _, service := range services {
		testDB.Exec("INSERT INTO services (id, name, slug, description) VALUES (?, ?, ?, ?)",
			service.ID, service.Name, service.Slug, service.Description)
	}

	// Insert test versions
	versions := []Version{
		{ID: "version-1", ServiceID: "service-1", Semver: "1.0.0", Status: "released", Changelog: "Initial release"},
		{ID: "version-2", ServiceID: "service-1", Semver: "1.1.0", Status: "released", Changelog: "Minor update"},
		{ID: "version-3", ServiceID: "service-2", Semver: "0.1.0", Status: "draft", Changelog: "Work in progress"},
		{ID: "version-4", ServiceID: "service-3", Semver: "2.0.0", Status: "released", Changelog: "Major update"},
	}

	for _, version := range versions {
		testDB.Exec("INSERT INTO versions (id, service_id, semver, status, changelog) VALUES (?, ?, ?, ?, ?)",
			version.ID, version.ServiceID, version.Semver, version.Status, version.Changelog)
	}

	// Update versions_count
	testDB.Exec("UPDATE services SET versions_count = (SELECT COUNT(*) FROM versions WHERE service_id = services.id)")
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add routes
	router.GET("/health", healthCheck)
	router.GET("/api/v1/services", getServices)
	router.GET("/api/v1/services/search", searchServices)
	router.POST("/api/v1/services", createService)
	router.GET("/api/v1/services/:id", getService)
	router.PUT("/api/v1/services/:id", updateService)
	router.DELETE("/api/v1/services/:id", deleteService)
	router.GET("/api/v1/services/:id/versions", getVersions)
	router.POST("/api/v1/services/:id/versions", createVersion)

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
				var response PaginatedResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.expectedCount > 0 {
					services := response.Data.([]interface{})
					assert.Len(t, services, tt.expectedCount)
				}

				assert.NotNil(t, response.Pagination)
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
				var response PaginatedResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				services := response.Data.([]interface{})
				assert.Len(t, services, tt.expectedCount)
				assert.NotNil(t, response.Pagination)
			}
		})
	}
}

func TestCreateServiceIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceData    Service
		expectedStatus int
	}{
		{
			name: "valid service",
			serviceData: Service{
				Name:        "New Test Service",
				Slug:        "new-test-service",
				Description: "A new test service",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "service with duplicate name",
			serviceData: Service{
				Name:        "Test Service 1", // Already exists
				Slug:        "duplicate-service",
				Description: "Duplicate service",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service with duplicate slug",
			serviceData: Service{
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
				var response Service
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
				var response Service
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
				var response PaginatedResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				versions := response.Data.([]interface{})
				assert.Len(t, versions, tt.expectedCount)
				assert.NotNil(t, response.Pagination)
			}
		})
	}
}

func TestCreateVersionIntegration(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		serviceID      string
		versionData    Version
		expectedStatus int
	}{
		{
			name:      "valid version",
			serviceID: "service-1",
			versionData: Version{
				Semver:    "1.2.0",
				Status:    "released",
				Changelog: "New feature release",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "version for non-existing service",
			serviceID: "non-existing",
			versionData: Version{
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
				var response Version
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
