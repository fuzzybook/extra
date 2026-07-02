CREATE TABLE welcome_tokens (
    token UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX welcome_tokens_user_id_idx ON welcome_tokens (user_id);
CREATE INDEX welcome_tokens_email_idx ON welcome_tokens (email);
