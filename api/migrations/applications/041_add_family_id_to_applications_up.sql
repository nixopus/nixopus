ALTER TABLE applications 
ADD COLUMN family_id UUID DEFAULT NULL;
CREATE INDEX idx_applications_family_id ON applications(family_id);

