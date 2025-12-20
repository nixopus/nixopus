DROP INDEX IF EXISTS idx_extensions_parent_extension_id;
ALTER TABLE extensions DROP COLUMN IF EXISTS parent_extension_id;

