-- Rollback Article Storage References
-- Remove storage reference columns added in the up migration

ALTER TABLE article DROP COLUMN IF EXISTS storage_backend;
ALTER TABLE article DROP COLUMN IF EXISTS storage_key;