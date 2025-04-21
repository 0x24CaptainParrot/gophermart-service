-- +goose Up
-- +goose StatementBegin
CREATE TABLE balance (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    current NUMERIC(20, 4) NOT NULL DEFAULT 0,
    withdrawn NUMERIC(20, 4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balance;
-- +goose StatementEnd
