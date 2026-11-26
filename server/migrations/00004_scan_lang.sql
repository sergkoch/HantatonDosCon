-- +goose Up
ALTER TABLE scans ADD COLUMN IF NOT EXISTS lang text NOT NULL DEFAULT 'ru';

-- +goose Down
ALTER TABLE scans DROP COLUMN IF EXISTS lang;
