-- Emails are compared case-insensitively: normalize stored values and make
-- the uniqueness constraint case-insensitive too.
UPDATE user_profiles SET email = lower(btrim(email)) WHERE email <> lower(btrim(email));
DROP INDEX IF EXISTS user_profiles_email_idx;
CREATE UNIQUE INDEX IF NOT EXISTS user_profiles_email_lower_idx ON user_profiles (lower(email));
