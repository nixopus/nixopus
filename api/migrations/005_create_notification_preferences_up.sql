CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

CREATE TABLE preference_item (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    preference_id UUID NOT NULL REFERENCES notification_preferences(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE smtp_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) NOT NULL,
    security VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT false,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_preference_item_preference_id ON preference_item(preference_id);
CREATE INDEX idx_preference_item_category ON preference_item(category);
CREATE INDEX idx_preference_item_type ON preference_item(type);
CREATE INDEX idx_smtp_configs_user_id ON smtp_configs(user_id);
CREATE INDEX idx_smtp_configs_is_active ON smtp_configs(is_active);

INSERT INTO notification_preferences (id, user_id)
SELECT uuid_generate_v4(), id
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM notification_preferences WHERE user_id = users.id
);

CREATE OR REPLACE FUNCTION seed_default_preference_items()
RETURNS void AS $$
DECLARE
    pref_record RECORD;
BEGIN
    FOR pref_record IN SELECT * FROM notification_preferences LOOP
        IF NOT EXISTS (SELECT 1 FROM preference_item WHERE preference_id = pref_record.id) THEN
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'activity', 'team-updates', true);
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'security', 'login-alerts', true);
            
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'security', 'password-changes', true);
            
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'security', 'security-alerts', true);

            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'update', 'product-updates', true);
            
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'update', 'newsletter', false);
            
            INSERT INTO preference_item (id, preference_id, category, type, enabled)
            VALUES (uuid_generate_v4(), pref_record.id, 'update', 'marketing', false);
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

SELECT seed_default_preference_items();

DROP FUNCTION seed_default_preference_items();