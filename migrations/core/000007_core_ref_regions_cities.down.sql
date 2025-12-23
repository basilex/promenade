-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS core_cities CASCADE;
DROP TABLE IF EXISTS core_regions CASCADE;

-- +goose StatementEnd
