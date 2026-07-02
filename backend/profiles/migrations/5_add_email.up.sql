ALTER TABLE user_profiles ADD COLUMN email TEXT NOT NULL DEFAULT '';

-- Existing rows predate the email column; give each a unique placeholder
-- derived from its user ID so the uniqueness constraint below can apply.
UPDATE user_profiles SET email = user_id || '@example.invalid' WHERE email = '';

CREATE UNIQUE INDEX user_profiles_email_idx ON user_profiles (email);
