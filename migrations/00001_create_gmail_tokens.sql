-- +goose Up
-- +goose StatementBegin

CREATE TABLE gmail_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_email TEXT NOT NULL UNIQUE,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT,
    expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
