ALTER TABLE applications ADD COLUMN IF NOT EXISTS compose_file_url TEXT DEFAULT '';
ALTER TABLE applications ADD COLUMN IF NOT EXISTS compose_file_content TEXT DEFAULT '';

