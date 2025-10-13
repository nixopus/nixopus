DROP INDEX IF EXISTS idx_extensions_extension_type;

ALTER TABLE extensions
    DROP COLUMN IF EXISTS extension_type;

DROP TYPE IF EXISTS extension_type;

