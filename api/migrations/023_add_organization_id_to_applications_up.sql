ALTER TABLE applications ADD COLUMN organization_id UUID;
UPDATE applications SET organization_id = (SELECT id FROM organizations LIMIT 1);
ALTER TABLE applications ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE applications ADD CONSTRAINT fk_applications_organization FOREIGN KEY (organization_id) REFERENCES organizations(id); 