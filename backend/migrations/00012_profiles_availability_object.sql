-- +goose Up
-- availability_slots по контракту — объект {"mon": [19, 20], ...}, а не массив.
-- Старое значение по умолчанию '[]' превращаем в NULL (не заполнено).
ALTER TABLE profiles ALTER COLUMN availability_slots DROP DEFAULT;
UPDATE profiles SET availability_slots = NULL WHERE jsonb_typeof(availability_slots) <> 'object';
ALTER TABLE profiles ADD CONSTRAINT profiles_availability_is_object
    CHECK (availability_slots IS NULL OR jsonb_typeof(availability_slots) = 'object');

-- +goose Down
ALTER TABLE profiles DROP CONSTRAINT profiles_availability_is_object;
ALTER TABLE profiles ALTER COLUMN availability_slots SET DEFAULT '[]';
