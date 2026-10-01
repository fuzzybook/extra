CREATE TABLE promo (
    id BIGSERIAL PRIMARY KEY,
    position INT NOT NULL CHECK (position IN (0, 1, 2)),
    item JSONB NOT NULL DEFAULT '[]'::jsonb,
    active_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
