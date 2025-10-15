CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ BEGIN
    CREATE TYPE extension_category AS ENUM (
        'Security', 'Containers', 'Database', 'Web Server', 
        'Maintenance', 'Monitoring', 'Storage', 'Network', 
        'Development', 'Other'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE validation_status AS ENUM (
        'not_validated', 'valid', 'invalid'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE execution_status AS ENUM (
        'pending', 'running', 'completed', 'failed'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS extensions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    extension_id VARCHAR(50) UNIQUE NOT NULL,
    
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL CHECK (LENGTH(description) BETWEEN 10 AND 500),
    author VARCHAR(50) NOT NULL,
    icon VARCHAR(10) NOT NULL,
    category extension_category NOT NULL,
    
    version VARCHAR(20),
    is_verified BOOLEAN NOT NULL DEFAULT false,
    
    yaml_content TEXT NOT NULL,
    parsed_content JSONB NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    
    validation_status validation_status DEFAULT 'not_validated',
    validation_errors JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT valid_extension_id CHECK (extension_id ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$'),
    CONSTRAINT valid_version CHECK (version IS NULL OR version ~ '^\d+\.\d+\.\d+(-[a-zA-Z0-9\-]+)?$')
);

CREATE TABLE IF NOT EXISTS extension_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    extension_id UUID REFERENCES extensions(id) ON DELETE CASCADE,
    
    variable_name VARCHAR(100) NOT NULL,
    variable_type VARCHAR(20) NOT NULL,
    description TEXT,
    default_value JSONB,
    is_required BOOLEAN DEFAULT false,
    validation_pattern VARCHAR(500),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(extension_id, variable_name),
    CONSTRAINT valid_variable_name CHECK (variable_name ~ '^[a-zA-Z_][a-zA-Z0-9_]*$'),
    CONSTRAINT valid_variable_type CHECK (variable_type IN ('string', 'integer', 'boolean', 'array'))
);

CREATE TABLE IF NOT EXISTS extension_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    extension_id UUID REFERENCES extensions(id) ON DELETE CASCADE,
    
    server_hostname VARCHAR(255),
    variable_values JSONB,
    
    status execution_status DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    exit_code INTEGER,
    error_message TEXT,
    execution_log TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS execution_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID REFERENCES extension_executions(id) ON DELETE CASCADE,
    
    step_name VARCHAR(200) NOT NULL,
    phase VARCHAR(20) NOT NULL,
    step_order INTEGER NOT NULL,
    
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    status execution_status DEFAULT 'pending',
    
    exit_code INTEGER,
    output TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_phase CHECK (phase IN ('pre_install', 'install', 'post_install', 'run', 'validate'))
);

CREATE INDEX IF NOT EXISTS idx_extensions_category ON extensions(category);
CREATE INDEX IF NOT EXISTS idx_extensions_verified ON extensions(is_verified);
CREATE INDEX IF NOT EXISTS idx_extensions_validation_status ON extensions(validation_status);
CREATE INDEX IF NOT EXISTS idx_extensions_created ON extensions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_extensions_extension_id ON extensions(extension_id);
CREATE INDEX IF NOT EXISTS idx_extensions_deleted_at ON extensions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_extension_variables_extension ON extension_variables(extension_id);
CREATE INDEX IF NOT EXISTS idx_extension_executions_extension ON extension_executions(extension_id);
CREATE INDEX IF NOT EXISTS idx_extension_executions_status ON extension_executions(status);
CREATE INDEX IF NOT EXISTS idx_execution_steps_execution ON execution_steps(execution_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_extensions_updated_at
  BEFORE UPDATE ON extensions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
