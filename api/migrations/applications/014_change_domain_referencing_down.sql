ALTER TABLE applications
ALTER COLUMN domain_id TYPE UUID USING domain_id::UUID;

ALTER TABLE applications
ADD CONSTRAINT fk_applications_domain_id 
FOREIGN KEY (domain_id) 
REFERENCES domains(id);

ALTER TABLE applications
RENAME COLUMN domain TO domain_id;