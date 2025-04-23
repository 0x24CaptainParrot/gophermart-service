-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number INTEGER NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual NUMERIC(20, 2) NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
