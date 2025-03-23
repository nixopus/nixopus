ALTER TABLE smtp_configs ADD COLUMN organization_id UUID NOT NULL;
DELETE FROM smtp_configs;
ALTER TABLE smtp_configs ADD CONSTRAINT fk_smtp_configs_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;