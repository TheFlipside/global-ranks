ALTER TABLE users ADD COLUMN secret_token VARCHAR(64) NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS sessions (
    id         UUID PRIMARY KEY,
    user_uuid  UUID NOT NULL REFERENCES users(uuid),
    game_id    INTEGER NOT NULL REFERENCES games(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used       BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_uuid, game_id);
