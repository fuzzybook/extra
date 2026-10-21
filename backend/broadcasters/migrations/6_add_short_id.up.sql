-- 6_add_short_id.up.sql
-- Aggiunge short_id: identificativo breve e unico di 6 caratteri [0-9a-z],
-- generato dal database alla creazione di ogni broadcaster.

-- Genera uno short_id non ancora usato.
CREATE OR REPLACE FUNCTION broadcasters_new_short_id() RETURNS TEXT AS $$
DECLARE
    alphabet CONSTANT TEXT := '0123456789abcdefghijklmnopqrstuvwxyz';
    candidate TEXT;
BEGIN
    LOOP
        candidate := '';
        FOR i IN 1..6 LOOP
            candidate := candidate || substr(alphabet, 1 + floor(random() * 36)::int, 1);
        END LOOP;
        EXIT WHEN NOT EXISTS (SELECT 1 FROM broadcasters WHERE short_id = candidate);
    END LOOP;
    RETURN candidate;
END;
$$ LANGUAGE plpgsql VOLATILE;

ALTER TABLE broadcasters ADD COLUMN short_id TEXT;

-- Record esistenti: un UPDATE per riga, così ogni nuovo valore vede quelli
-- assegnati prima e non può ripeterli.
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT id FROM broadcasters WHERE short_id IS NULL ORDER BY created_at, id LOOP
        UPDATE broadcasters SET short_id = broadcasters_new_short_id() WHERE id = r.id;
    END LOOP;
END;
$$;

ALTER TABLE broadcasters
    ALTER COLUMN short_id SET NOT NULL,
    ADD CONSTRAINT broadcasters_short_id_key UNIQUE (short_id),
    ADD CONSTRAINT broadcasters_short_id_format CHECK (short_id ~ '^[0-9a-z]{6}$');

-- Nuovi record: short_id assegnato automaticamente se non indicato.
CREATE OR REPLACE FUNCTION broadcasters_set_short_id() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.short_id IS NULL THEN
        NEW.short_id := broadcasters_new_short_id();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER broadcasters_short_id_before_insert
    BEFORE INSERT ON broadcasters
    FOR EACH ROW EXECUTE FUNCTION broadcasters_set_short_id();
