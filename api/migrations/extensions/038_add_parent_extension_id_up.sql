ALTER TABLE extensions
  ADD COLUMN IF NOT EXISTS parent_extension_id UUID NULL REFERENCES extensions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_extensions_parent_extension_id ON extensions(parent_extension_id);

