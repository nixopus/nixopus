ALTER TABLE applications
ALTER COLUMN domain_id TYPE TEXT;

ALTER TABLE applications
RENAME COLUMN domain_id TO domain;