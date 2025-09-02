-- Add new resource types to the audit_resource_type enum
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'notification';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'feature_flag';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'file_manager';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'container';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'audit';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'terminal';
ALTER TYPE audit_resource_type ADD VALUE IF NOT EXISTS 'integration';
