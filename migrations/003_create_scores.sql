CREATE TABLE IF NOT EXISTS scores (
    id           BIGSERIAL PRIMARY KEY,
    user_uuid    UUID NOT NULL REFERENCES users(uuid),
    game_id      INTEGER NOT NULL REFERENCES games(id),
    score        BIGINT NOT NULL CHECK (score >= 0),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scores_leaderboard ON scores (game_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_scores_user_game ON scores (user_uuid, game_id);
