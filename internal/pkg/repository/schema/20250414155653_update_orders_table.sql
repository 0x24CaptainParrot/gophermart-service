-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE orders
ALTER COLUMN accrual SET DEFAULT 0;
UPDATE orders SET accrual = 0 WHERE accrual IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
ALTER COLUMN accrual DROP DEFAULT;

ALTER TABLE orders
DROP COLUMN updated_at;
-- +goose StatementEnd
