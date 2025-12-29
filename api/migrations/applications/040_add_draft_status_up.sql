-- Add 'draft' value to the status enum type for applications that are saved but not yet deployed
ALTER TYPE status ADD VALUE IF NOT EXISTS 'draft';

