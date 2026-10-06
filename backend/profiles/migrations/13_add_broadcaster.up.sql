ALTER TABLE user_profiles ADD COLUMN broadcaster UUID;
CREATE INDEX idx_user_profiles_broadcaster ON user_profiles(broadcaster);
