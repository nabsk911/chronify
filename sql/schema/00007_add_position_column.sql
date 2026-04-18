-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN position INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN position;
-- +goose StatementEnd
