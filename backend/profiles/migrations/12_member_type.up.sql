ALTER TABLE user_profiles ADD COLUMN member_type TEXT NOT NULL DEFAULT 'visitor';
UPDATE user_profiles SET member_type = 'broadcaster' WHERE is_artist = TRUE;
ALTER TABLE user_profiles DROP COLUMN is_artist;
