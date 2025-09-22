package models

// Service represents a service entity in the system
type Service struct {
	ID            string `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Slug          string `json:"slug" db:"slug"`
	Description   string `json:"description" db:"description"`
	CreatedAt     string `json:"created_at" db:"created_at"`
	UpdatedAt     string `json:"updated_at" db:"updated_at"`
	VersionsCount int    `json:"versions_count" db:"versions_count"`
}
