-- +goose Up
-- Заявка, созданную куратором при ручном подборе (POST /curator/requests),
-- не учитываем в лимите 3 активных заявок ученика. Поля нет в ER-диаграмме,
-- но без него эти заявки не отличить от заявок самого ученика.
ALTER TABLE requests
    ADD COLUMN suggested_by UUID REFERENCES users (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE requests DROP COLUMN suggested_by;
