ALTER TABLE applications
ADD COLUMN server_id UUID REFERENCES servers(id) ON DELETE SET NULL;

CREATE INDEX idx_applications_server_id ON applications(server_id) WHERE server_id IS NOT NULL;
CREATE INDEX idx_applications_org_server ON applications(organization_id, server_id) WHERE server_id IS NOT NULL;
