ALTER TABLE applications ADD COLUMN dockerfile_path TEXT;

UPDATE applications SET dockerfile_path = 'Dockerfile' WHERE dockerfile_path IS NULL;

ALTER TABLE applications ALTER COLUMN dockerfile_path SET NOT NULL;

ALTER TABLE applications ALTER COLUMN dockerfile_path SET DEFAULT 'Dockerfile';