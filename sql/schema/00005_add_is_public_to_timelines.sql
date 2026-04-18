-- +goose Up
-- +goose StatementBegin
ALTER TABLE timelines ADD COLUMN is_public BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE timelines DROP COLUMN is_public;
-- +goose StatementEnd
