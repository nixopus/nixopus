DO $$ BEGIN
    CREATE TYPE extension_type AS ENUM ('install', 'run');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE extensions
    ADD COLUMN IF NOT EXISTS extension_type extension_type NOT NULL DEFAULT 'run';

CREATE INDEX IF NOT EXISTS idx_extensions_extension_type ON extensions(extension_type);

