-- +goose Up
-- dog_id and handler_id are nullable so existing entries (which only carry
-- denormalized handler_name / dog_name strings) keep working. New entries
-- created through the registration flow populate both. The string columns
-- stay as a print-program snapshot so historical results still display
-- correctly if a dog is later renamed or transferred.
ALTER TABLE entries ADD COLUMN dog_id INTEGER REFERENCES dogs(id);
ALTER TABLE entries ADD COLUMN handler_id INTEGER REFERENCES competitors(id);

CREATE INDEX idx_entries_dog_id ON entries(dog_id);
CREATE INDEX idx_entries_handler_id ON entries(handler_id);

-- +goose Down
DROP INDEX IF EXISTS idx_entries_handler_id;
DROP INDEX IF EXISTS idx_entries_dog_id;
ALTER TABLE entries DROP COLUMN handler_id;
ALTER TABLE entries DROP COLUMN dog_id;
