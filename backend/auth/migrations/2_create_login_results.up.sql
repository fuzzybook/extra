CREATE TABLE login_results (
    request_id UUID PRIMARY KEY,
    ok BOOLEAN NOT NULL,
    token UUID,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX login_results_created_at_idx ON login_results (created_at);
