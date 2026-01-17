CREATE TABLE IF NOT EXISTS application_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL,
    domain TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    CONSTRAINT unique_domain_per_application UNIQUE (application_id, domain)
);

CREATE INDEX IF NOT EXISTS idx_application_domains_application_id ON application_domains(application_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_application_domains_domain_unique ON application_domains(domain);

-- Migrate existing domain data from applications table
-- Use ON CONFLICT to skip duplicate domains (domains must be globally unique)
-- If multiple applications have the same domain, only the first one encountered will be migrated
INSERT INTO application_domains (application_id, domain, created_at)
SELECT 
    id as application_id,
    domain,
    created_at
FROM applications
WHERE domain IS NOT NULL AND domain != ''
ON CONFLICT (domain) DO NOTHING;
