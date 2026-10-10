-- 5_chatlog_room_text.up.sql
-- Modifica la colonna "room" da UUID a TEXT per supportare qualsiasi identificatore di stanza (es. nomi LiveKit, slug, broadcast-uuid).

ALTER TABLE chatlog ALTER COLUMN room TYPE TEXT;
