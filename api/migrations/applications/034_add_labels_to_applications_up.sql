ALTER TABLE applications ADD COLUMN IF NOT EXISTS labels TEXT[] DEFAULT '{}';
CREATE INDEX IF NOT EXISTS idx_applications_labels ON applications USING GIN(labels);
