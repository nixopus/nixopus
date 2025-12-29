DROP INDEX IF EXISTS idx_applications_family_id;
ALTER TABLE applications DROP COLUMN IF EXISTS family_id;
