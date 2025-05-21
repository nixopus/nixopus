ALTER TABLE application_deployment 
ADD COLUMN IF NOT EXISTS container_id TEXT;

ALTER TABLE application_deployment 
ADD COLUMN IF NOT EXISTS container_name TEXT;

ALTER TABLE application_deployment 
ADD COLUMN IF NOT EXISTS container_image TEXT;

ALTER TABLE application_deployment 
ADD COLUMN IF NOT EXISTS container_status TEXT;