-- +goose Up
-- Sex is optional (empty string when unrecorded) so a dog can exist in the
-- registry before its sex is logged. CHECK keeps the column to the known set.
ALTER TABLE dogs ADD COLUMN sex TEXT NOT NULL DEFAULT '' CHECK (sex IN ('male', 'female', ''));

-- +goose Down
ALTER TABLE dogs DROP COLUMN sex;
