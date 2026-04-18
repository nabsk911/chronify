-- +goose Up
-- +goose StatementBegin
CREATE TABLE bookmarks (
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
timeline_id UUID NOT NULL REFERENCES timelines(id) ON DELETE CASCADE,
PRIMARY KEY (user_id, timeline_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE bookmarks;
-- +goose StatementEnd
