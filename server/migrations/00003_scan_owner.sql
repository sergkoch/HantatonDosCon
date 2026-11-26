-- +goose Up
ALTER TABLE scans ADD COLUMN IF NOT EXISTS owner_id text NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS scans_owner_created_idx ON scans (owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS scans_owner_url_created_idx ON scans (owner_id, url, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS scans_owner_url_created_idx;
DROP INDEX IF EXISTS scans_owner_created_idx;
ALTER TABLE scans DROP COLUMN IF EXISTS owner_id;
