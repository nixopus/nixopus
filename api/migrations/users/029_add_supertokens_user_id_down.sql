DROP INDEX IF EXISTS idx_users_supertokens_user_id;
ALTER TABLE users
DROP COLUMN IF EXISTS supertokens_user_id;

ALTER TABLE users
ALTER COLUMN password SET NOT NULL;
