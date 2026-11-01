-- Stato delle stanze broadcast aggiornato dai webhook di LiveKit: una riga per
-- stanza aperta, cancellata alla chiusura. Lo legge ListLive per il sito pubblico.

CREATE TABLE live_status (
    room_name TEXT PRIMARY KEY,
    room_sid TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    -- Proprietario collegato alla stanza, camera e microfono accesi (non in muto).
    owner_connected BOOLEAN NOT NULL DEFAULT FALSE,
    video BOOLEAN NOT NULL DEFAULT FALSE,
    audio BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_live_status_owner ON live_status(owner_id);
