CREATE TABLE personal_setup (
    user_id UUID PRIMARY KEY REFERENCES user_profiles (user_id) ON DELETE CASCADE,
    can_broadcast BOOLEAN NOT NULL DEFAULT FALSE,
    room_name TEXT NOT NULL DEFAULT '',
    CONSTRAINT personal_setup_broadcast_room_check CHECK (
        (NOT can_broadcast AND room_name = '') OR
        (can_broadcast AND octet_length(room_name) BETWEEN 1 AND 128 AND btrim(room_name) <> '')
    )
);
