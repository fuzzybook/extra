-- 4_add_enabled_used.up.sql
-- Aggiunge le colonne "enabled" e "used" alla tabella broadcasters.
-- Queste colonne sono usate dal codice Go per indicare lo stato del profilo.

ALTER TABLE broadcasters
    ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN used    BOOLEAN NOT NULL DEFAULT FALSE;

-- Indici facoltativi per velocizzare le query su queste colonne
CREATE INDEX idx_broadcasters_enabled ON broadcasters(enabled);
CREATE INDEX idx_broadcasters_used    ON broadcasters(used);
