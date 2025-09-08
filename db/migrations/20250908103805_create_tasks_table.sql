-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    words JSONB,
    result JSONB,
    file_path TEXT,
    case_sensitive BOOLEAN NOT NULL,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    processing_time_ms BIGINT,
    groups_count INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd
