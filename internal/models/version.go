package models

// Version represents a version of a service
type Version struct {
	ID        string `json:"id" db:"id"`
	ServiceID string `json:"service_id" db:"service_id"`
	Semver    string `json:"semver" db:"semver"`
	Status    string `json:"status" db:"status"`
	Changelog string `json:"changelog" db:"changelog"`
	CreatedAt string `json:"created_at" db:"created_at"`
}
