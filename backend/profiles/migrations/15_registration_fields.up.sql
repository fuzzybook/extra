-- 15_registration_fields.up.sql
-- Dati raccolti dal modulo di registrazione pubblico: lingua preferita,
-- iscrizione alla newsletter, accettazione dei termini e tipo di membro
-- richiesto (chi chiede "broadcaster" resta visitor finché un admin non lo
-- conferma). Il telefono va nei dati personali.
ALTER TABLE user_profiles
    ADD COLUMN language TEXT NOT NULL DEFAULT 'en',
    ADD COLUMN newsletter BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN terms_accepted_at TIMESTAMPTZ,
    ADD COLUMN requested_member_type TEXT NOT NULL DEFAULT '';

ALTER TABLE personal_data ADD COLUMN phone TEXT NOT NULL DEFAULT '';
