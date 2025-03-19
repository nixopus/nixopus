ALTER TABLE application_deployment 
DROP COLUMN IF EXISTS container_id;

ALTER TABLE application_deployment 
DROP COLUMN IF EXISTS container_name;

ALTER TABLE application_deployment 
DROP COLUMN IF EXISTS container_image;

ALTER TABLE application_deployment 
DROP COLUMN IF EXISTS container_status;