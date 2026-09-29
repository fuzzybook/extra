CREATE TABLE siteconfig (
    id BIGSERIAL PRIMARY KEY,
    extrsite JSONB NOT NULL,
    start_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    description TEXT NOT NULL DEFAULT '',
    CONSTRAINT siteconfig_extrsite_object CHECK (jsonb_typeof(extrsite) = 'object')
);
