-- +goose Up
ALTER TABLE findings ADD COLUMN IF NOT EXISTS title_alt text NOT NULL DEFAULT '';
ALTER TABLE findings ADD COLUMN IF NOT EXISTS recommendation_alt text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE findings DROP COLUMN IF EXISTS recommendation_alt;
ALTER TABLE findings DROP COLUMN IF EXISTS title_alt;
