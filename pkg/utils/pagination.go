package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yashjain/konnect/pkg/types"
)

// GetPaginationParams extracts and validates pagination parameters from request
func GetPaginationParams(c *gin.Context) types.PaginationParams {
	params := types.PaginationParams{
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

// GetSearchParams extracts and validates search parameters from request
func GetSearchParams(c *gin.Context) types.SearchParams {
	params := types.SearchParams{
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

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, pageSize, total int) types.Pagination {
	totalPages := (total + pageSize - 1) / pageSize // Ceiling division

	return types.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
