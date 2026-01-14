ALTER TABLE extensions
    ADD COLUMN IF NOT EXISTS featured BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_extensions_featured ON extensions(featured);
