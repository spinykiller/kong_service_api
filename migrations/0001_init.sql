-- +goose Up
CREATE TABLE services (
  id            CHAR(36)     NOT NULL,
  name          VARCHAR(255) NOT NULL,
  slug          VARCHAR(255) NOT NULL,
  description   TEXT NULL,
  created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_services_name (name),
  UNIQUE KEY uq_services_slug (slug),
  FULLTEXT KEY ft_services_name_desc (name, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE versions (
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

-- Optional denormalized count (comment out if not desired)
ALTER TABLE services ADD COLUMN versions_count INT NOT NULL DEFAULT 0;

-- +goose Down
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS services;
