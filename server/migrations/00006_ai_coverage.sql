-- +goose Up
ALTER TABLE scans ADD COLUMN IF NOT EXISTS ai jsonb;

-- +goose Down
ALTER TABLE scans DROP COLUMN IF EXISTS ai;
