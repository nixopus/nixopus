ALTER TABLE users
ADD COLUMN supertokens_user_id TEXT UNIQUE;

CREATE INDEX idx_users_supertokens_user_id ON users(supertokens_user_id);

-- Make password field optional for  users since we are using Supertokens
ALTER TABLE users
ALTER COLUMN password DROP NOT NULL;
