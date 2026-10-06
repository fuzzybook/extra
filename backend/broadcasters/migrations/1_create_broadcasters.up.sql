CREATE TABLE broadcasters (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    category_color TEXT NOT NULL DEFAULT '#6366f1',
    bio TEXT NOT NULL DEFAULT '',
    certified BOOLEAN NOT NULL DEFAULT FALSE,
    availability TEXT NOT NULL DEFAULT 'available',
    phone TEXT NOT NULL DEFAULT '',
    services JSONB NOT NULL DEFAULT '[]'::jsonb,
    languages JSONB NOT NULL DEFAULT '[]'::jsonb,
    attributes JSONB NOT NULL DEFAULT '[]'::jsonb,
    hours JSONB NOT NULL DEFAULT '[]'::jsonb,
    photos JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_broadcasters_user_id ON broadcasters(user_id);
CREATE INDEX idx_broadcasters_category ON broadcasters(category);
CREATE INDEX idx_broadcasters_availability ON broadcasters(availability);
