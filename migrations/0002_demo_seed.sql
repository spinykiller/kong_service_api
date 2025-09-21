-- +goose Up
INSERT INTO services (id, name, slug, description) VALUES
  (UUID(), 'Locate Us', 'locate-us', 'Store locator'),
  (UUID(), 'Collect Money', 'collect-money', 'Payments collection'),
  (UUID(), 'Notifications', 'notifications', 'Send push/email');

INSERT INTO versions (id, service_id, semver, status, changelog) VALUES
  (UUID(), (SELECT id FROM services WHERE slug='notifications'), '1.0.0', 'released', 'Initial release'),
  (UUID(), (SELECT id FROM services WHERE slug='notifications'), '1.1.0', 'released', 'Minor improvements'),
  (UUID(), (SELECT id FROM services WHERE slug='collect-money'), '0.1.0', 'draft', 'WIP');

-- Update versions_count manually since we removed triggers
UPDATE services SET versions_count = (
  SELECT COUNT(*) FROM versions WHERE service_id = services.id
);

-- +goose Down
DELETE FROM versions;
DELETE FROM services;
