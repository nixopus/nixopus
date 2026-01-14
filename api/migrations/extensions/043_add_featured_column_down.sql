DROP INDEX IF EXISTS idx_extensions_featured;

ALTER TABLE extensions
    DROP COLUMN IF EXISTS featured;
