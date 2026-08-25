CREATE TABLE IF NOT EXISTS rules (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    backend TEXT NOT NULL,
    percent INT NOT NULL CHECK (percent >= 0 AND percent <= 100)
    created_at TIMESTAMPTZ DEFAULT NOW()
    updated_at TIMESTAMPTZ DEFAULT NOW()
);