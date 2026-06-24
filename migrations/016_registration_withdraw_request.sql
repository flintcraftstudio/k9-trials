-- +goose Up
-- Withdrawal of an accepted registration is a competitor request that an
-- admin confirms (Q1). withdraw_requested_at marks the pending request; the
-- terminal 'withdrawn' status (already in the CHECK) is set on admin confirm,
-- retaining the entry row and its entry_number for audit.
ALTER TABLE registrations ADD COLUMN withdraw_requested_at DATETIME;

-- +goose Down
ALTER TABLE registrations DROP COLUMN withdraw_requested_at;
