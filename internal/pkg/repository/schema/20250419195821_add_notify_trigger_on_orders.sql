-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION notify_new_order()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('order_notifications', NEW.number::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER order_notification
AFTER INSERT ON orders 
FOR EACH ROW EXECUTE FUNCTION notify_new_order();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS order_notification ON orders;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS notify_new_order()
-- +goose StatementEnd
