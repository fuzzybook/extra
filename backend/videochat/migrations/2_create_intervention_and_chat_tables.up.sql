-- 2_create_intervention_and_chat_tables.up.sql
-- Tabella per le richieste di intervento (alzata di mano) dei viewer
CREATE TABLE videochat_speak_requests (
    id UUID PRIMARY KEY,
    room_name TEXT NOT NULL,
    requester_user_id TEXT NOT NULL,
    requester_livekit_identity TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'expired', 'revoked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    decided_at TIMESTAMPTZ,
    decided_by_user_id TEXT
);

-- Indice per ricerca rapida delle richieste per stanza e stato
CREATE INDEX idx_speak_requests_room_status ON videochat_speak_requests(room_name, status, created_at);

-- Unicità parziale: un utente può avere al massimo UNA richiesta 'pending' attiva per stanza
CREATE UNIQUE INDEX idx_speak_requests_unique_pending 
ON videochat_speak_requests(room_name, requester_user_id) 
WHERE status = 'pending';

-- Tabella per lo slot unico di guest speaker nella stanza
CREATE TABLE videochat_speaker_slots (
    room_name TEXT PRIMARY KEY,
    speaker_user_id TEXT,
    speaker_livekit_identity TEXT,
    request_id UUID REFERENCES videochat_speak_requests(id) ON DELETE SET NULL,
    state TEXT NOT NULL CHECK (state IN ('free', 'reserved', 'active', 'releasing')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_expires_at TIMESTAMPTZ
);

-- Tabella per la messaggistica chat persistente della stanza
CREATE TABLE videochat_messages (
    id UUID PRIMARY KEY,
    room_name TEXT NOT NULL,
    sender_user_id TEXT NOT NULL,
    sender_name TEXT NOT NULL DEFAULT '',
    sender_avatar TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'text' CHECK (kind IN ('text', 'system')),
    body TEXT NOT NULL,
    client_message_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indice per paginazione stabile cronologica dei messaggi
CREATE INDEX idx_videochat_messages_room_created ON videochat_messages(room_name, created_at, id);

-- Unicità per idempotenza: impedisce l'inserimento duplicato dello stesso client_message_id
CREATE UNIQUE INDEX idx_videochat_messages_idempotency ON videochat_messages(room_name, sender_user_id, client_message_id);
