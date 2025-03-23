DELETE FROM domains;
ALTER TABLE domains ADD COLUMN organization_id uuid NOT NULL;
ALTER TABLE domains ADD CONSTRAINT fk_domains_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;