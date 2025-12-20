ALTER TABLE applications DROP COLUMN IF EXISTS labels;
DROP INDEX IF EXISTS idx_applications_labels;
