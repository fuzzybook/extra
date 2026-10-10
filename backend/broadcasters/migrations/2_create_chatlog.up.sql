CREATE TABLE chatlog (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    room UUID NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('ACTION','MESSAGE','ALLERT','CMD')),
    data TEXT NOT NULL
);

CREATE INDEX idx_chatlog_room ON chatlog(room);
