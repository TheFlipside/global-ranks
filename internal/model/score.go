package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Score represents a single score submission.
type Score struct {
	ID          int64     `json:"id"`
	UserUUID    uuid.UUID `json:"user_uuid"`
	GameID      int       `json:"game_id"`
	Score       int64     `json:"score"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// LeaderboardEntry represents one row in a leaderboard response.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	UUID        uuid.UUID `json:"uuid"`
	Username    string    `json:"username"`
	Score       int64     `json:"score"`
	AvatarURL   string    `json:"avatar_url"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// SubmitScore inserts a new score and returns the entry ID and the user's
// current rank for that game.
func SubmitScore(ctx context.Context, db *sql.DB, userUUID uuid.UUID, gameID int, score int64) (int64, int, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO scores (user_uuid, game_id, score)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		userUUID, gameID, score,
	).Scan(&id)
	if err != nil {
		return 0, 0, fmt.Errorf("submit score: %w", err)
	}

	var rank int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) + 1 FROM (
			SELECT user_uuid, MAX(score) AS best
			FROM scores WHERE game_id = $1
			GROUP BY user_uuid
			HAVING MAX(score) > $2
		) AS better`,
		gameID, score,
	).Scan(&rank)
	if err != nil {
		return id, 0, fmt.Errorf("compute rank: %w", err)
	}

	return id, rank, nil
}

// GetLeaderboard returns the top scores for a game, one entry per user (best score).
func GetLeaderboard(ctx context.Context, db *sql.DB, gameID, limit, offset int) ([]LeaderboardEntry, int, error) {
	var total int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_uuid) FROM scores WHERE game_id = $1`,
		gameID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count leaderboard: %w", err)
	}

	rows, err := db.QueryContext(ctx,
		`WITH best AS (
			SELECT user_uuid, MAX(score) AS score, MAX(submitted_at) AS submitted_at
			FROM scores
			WHERE game_id = $1
			GROUP BY user_uuid
		)
		SELECT b.user_uuid, u.username, b.score, b.submitted_at
		FROM best b
		JOIN users u ON u.uuid = b.user_uuid
		ORDER BY b.score DESC, b.submitted_at ASC
		LIMIT $2 OFFSET $3`,
		gameID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UUID, &e.Username, &e.Score, &e.SubmittedAt); err != nil {
			return nil, 0, fmt.Errorf("scan leaderboard row: %w", err)
		}
		e.Rank = rank
		e.AvatarURL = fmt.Sprintf("/api/v1/avatars/%s.png", e.UUID)
		rank++
		entries = append(entries, e)
	}

	return entries, total, rows.Err()
}
