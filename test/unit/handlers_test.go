package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yashjain/konnect/internal/handlers"
	"github.com/yashjain/konnect/internal/models"
	"github.com/yashjain/konnect/pkg/types"
	"github.com/yashjain/konnect/pkg/utils"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handlers.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGetPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryParams  string
		expectedPage int
		expectedSize int
	}{
		{
			name:         "default values",
			queryParams:  "",
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "custom values",
			queryParams:  "?page=2&page_size=5",
			expectedPage: 2,
			expectedSize: 5,
		},
		{
			name:         "only page",
			queryParams:  "?page=3",
			expectedPage: 3,
			expectedSize: 10,
		},
		{
			name:         "only page_size",
			queryParams:  "?page_size=20",
			expectedPage: 1,
			expectedSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				params := utils.GetPaginationParams(c)
				c.JSON(http.StatusOK, gin.H{
					"page":      params.Page,
					"page_size": params.PageSize,
				})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, float64(tt.expectedPage), response["page"])
			assert.Equal(t, float64(tt.expectedSize), response["page_size"])
		})
	}
}

func TestGetSearchParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		queryParams   string
		expectedQuery string
		expectedPage  int
		expectedSize  int
	}{
		{
			name:          "with search query",
			queryParams:   "?q=test&page=2&page_size=5",
			expectedQuery: "test",
			expectedPage:  2,
			expectedSize:  5,
		},
		{
			name:          "only query",
			queryParams:   "?q=notification",
			expectedQuery: "notification",
			expectedPage:  1,
			expectedSize:  10,
		},
		{
			name:          "empty query",
			queryParams:   "?q=",
			expectedQuery: "",
			expectedPage:  1,
			expectedSize:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				params := utils.GetSearchParams(c)
				c.JSON(http.StatusOK, gin.H{
					"query":     params.Query,
					"page":      params.Page,
					"page_size": params.PageSize,
				})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedQuery, response["query"])
			assert.Equal(t, float64(tt.expectedPage), response["page"])
			assert.Equal(t, float64(tt.expectedSize), response["page_size"])
		})
	}
}

func TestCalculatePagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		total    int
		expected types.Pagination
	}{
		{
			name:     "first page",
			page:     1,
			pageSize: 10,
			total:    25,
			expected: types.Pagination{
				Page:       1,
				PageSize:   10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name:     "middle page",
			page:     2,
			pageSize: 10,
			total:    25,
			expected: types.Pagination{
				Page:       2,
				PageSize:   10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    true,
			},
		},
		{
			name:     "last page",
			page:     3,
			pageSize: 10,
			total:    25,
			expected: types.Pagination{
				Page:       3,
				PageSize:   10,
				Total:      25,
				TotalPages: 3,
				HasNext:    false,
				HasPrev:    true,
			},
		},
		{
			name:     "exact page size",
			page:     1,
			pageSize: 10,
			total:    10,
			expected: types.Pagination{
				Page:       1,
				PageSize:   10,
				Total:      10,
				TotalPages: 1,
				HasNext:    false,
				HasPrev:    false,
			},
		},
		{
			name:     "empty results",
			page:     1,
			pageSize: 10,
			total:    0,
			expected: types.Pagination{
				Page:       1,
				PageSize:   10,
				Total:      0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CalculatePagination(tt.page, tt.pageSize, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceStruct(t *testing.T) {
	service := models.Service{
		ID:            "test-id",
		Name:          "Test Service",
		Slug:          "test-service",
		Description:   "A test service",
		CreatedAt:     "2023-01-01T00:00:00Z",
		UpdatedAt:     "2023-01-01T00:00:00Z",
		VersionsCount: 5,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(service)
	require.NoError(t, err)

	var unmarshaled models.Service
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, service, unmarshaled)
}

func TestVersionStruct(t *testing.T) {
	version := models.Version{
		ID:        "test-id",
		ServiceID: "service-id",
		Semver:    "1.0.0",
		Status:    "released",
		Changelog: "Initial release",
		CreatedAt: "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(version)
	require.NoError(t, err)

	var unmarshaled models.Version
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, version, unmarshaled)
}

func TestPaginatedResponseStruct(t *testing.T) {
	services := []models.Service{
		{ID: "1", Name: "Service 1"},
		{ID: "2", Name: "Service 2"},
	}

	pagination := types.Pagination{
		Page:       1,
		PageSize:   10,
		Total:      2,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	response := types.PaginatedResponse{
		Data:       services,
		Pagination: pagination,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled types.PaginatedResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, response.Pagination, unmarshaled.Pagination)
}
