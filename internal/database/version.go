package database

import (
	"log"

	"github.com/yashjain/konnect/internal/models"
	"github.com/yashjain/konnect/pkg/types"
)

// GetVersions retrieves paginated versions for a service
func GetVersions(serviceID string, params types.PaginationParams) ([]models.Version, int, error) {
	offset := (params.Page - 1) * params.PageSize

	// Get total count for this service
	var total int
	err := DB.QueryRow("SELECT COUNT(*) FROM versions WHERE service_id = ?", serviceID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated versions
	query := "SELECT id, service_id, semver, status, changelog, created_at FROM versions WHERE service_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	rows, err := DB.Query(query, serviceID, params.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var versions []models.Version
	for rows.Next() {
		var v models.Version
		err := rows.Scan(&v.ID, &v.ServiceID, &v.Semver, &v.Status, &v.Changelog, &v.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		versions = append(versions, v)
	}

	return versions, total, nil
}

// CreateVersion creates a new version for a service
func CreateVersion(version *models.Version) error {
	// Start a transaction to ensure atomicity
	tx, err := DB.Begin()
	if err != nil {
		return err
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
		return err
	}

	// Update the versions_count in the services table
	_, err = tx.Exec("UPDATE services SET versions_count = versions_count + 1 WHERE id = ?", version.ServiceID)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
