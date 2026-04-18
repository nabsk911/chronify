-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN media_name TEXT,
  ADD COLUMN media_type VARCHAR(255),
  ADD COLUMN media_url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
DROP COLUMN  media_name,
DROP COLUMN  media_type,
DROP COLUMN  media_url;
-- +goose StatementEnd
