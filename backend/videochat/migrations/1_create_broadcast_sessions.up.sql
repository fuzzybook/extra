-- 1_create_broadcast_sessions.up.sql
-- Tabella per registrare ogni apertura e chiusura delle trasmissioni nel servizio videochat.

CREATE TABLE broadcast_sessions (
    id BIGSERIAL PRIMARY KEY,
    owner_id TEXT NOT NULL,
    sid TEXT NOT NULL,
    name TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('OPEN', 'CLOSE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_broadcast_sessions_owner ON broadcast_sessions(owner_id);
CREATE INDEX idx_broadcast_sessions_sid ON broadcast_sessions(sid);
CREATE INDEX idx_broadcast_sessions_name ON broadcast_sessions(name);
CREATE INDEX idx_broadcast_sessions_created_at ON broadcast_sessions(created_at);
