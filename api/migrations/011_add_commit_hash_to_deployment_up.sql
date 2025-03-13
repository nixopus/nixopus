ALTER TABLE application_deployment 
ADD COLUMN IF NOT EXISTS commit_hash TEXT;