-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS core_payment_methods CASCADE;

-- +goose StatementEnd
