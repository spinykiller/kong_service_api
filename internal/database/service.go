package database

import (
	"log"

	"github.com/yashjain/konnect/internal/models"
	"github.com/yashjain/konnect/pkg/types"
)

// GetServices retrieves paginated services from the database
func GetServices(params types.PaginationParams) ([]models.Service, int, error) {
	offset := (params.Page - 1) * params.PageSize

	// Get total count
	var total int
	err := DB.QueryRow("SELECT COUNT(*) FROM services").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated services
	query := "SELECT id, name, slug, description, created_at, updated_at, versions_count FROM services ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := DB.Query(query, params.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.VersionsCount)
		if err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}

	return services, total, nil
}

// SearchServices performs full-text search on services
func SearchServices(params types.SearchParams) ([]models.Service, int, error) {
	offset := (params.Page - 1) * params.PageSize

	// Get total count for search results
	countQuery := "SELECT COUNT(*) FROM services WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)"
	var total int
	err := DB.QueryRow(countQuery, params.Query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated search results
	searchQuery := `
		SELECT id, name, slug, description, created_at, updated_at, versions_count 
		FROM services 
		WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)
		ORDER BY MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE) DESC, created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := DB.Query(searchQuery, params.Query, params.Query, params.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.VersionsCount)
		if err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}

	return services, total, nil
}

// CreateService creates a new service in the database
func CreateService(service *models.Service) error {
	_, err := DB.Exec("INSERT INTO services (id, name, slug, description) VALUES (?, ?, ?, ?)",
		service.ID, service.Name, service.Slug, service.Description)
	return err
}

// GetServiceByID retrieves a service by its ID
func GetServiceByID(id string) (*models.Service, error) {
	var service models.Service
	err := DB.QueryRow("SELECT id, name, slug, description, created_at, updated_at, versions_count FROM services WHERE id = ?", id).
		Scan(&service.ID, &service.Name, &service.Slug, &service.Description, &service.CreatedAt, &service.UpdatedAt, &service.VersionsCount)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// UpdateService updates a service in the database
func UpdateService(id string, service *models.Service) (int64, error) {
	result, err := DB.Exec("UPDATE services SET name = ?, slug = ?, description = ? WHERE id = ?",
		service.Name, service.Slug, service.Description, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

// DeleteService deletes a service from the database
func DeleteService(id string) (int64, error) {
	result, err := DB.Exec("DELETE FROM services WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}
