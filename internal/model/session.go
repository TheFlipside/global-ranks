package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session represents a game play session.
type Session struct {
	ID        uuid.UUID `json:"id"`
	UserUUID  uuid.UUID `json:"user_uuid"`
	GameID    int       `json:"game_id"`
	StartedAt time.Time `json:"started_at"`
	Used      bool      `json:"used"`
}

// CreateSession starts a new game session for the given user and game.
func CreateSession(ctx context.Context, db *sql.DB, userUUID uuid.UUID, gameID int) (*Session, error) {
	id := uuid.New()
	s := &Session{ID: id, UserUUID: userUUID, GameID: gameID}

	err := db.QueryRowContext(ctx,
		`INSERT INTO sessions (id, user_uuid, game_id)
		 VALUES ($1, $2, $3)
		 RETURNING started_at, used`,
		id, userUUID, gameID,
	).Scan(&s.StartedAt, &s.Used)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return s, nil
}

// GetSession retrieves a session by ID.
func GetSession(ctx context.Context, db *sql.DB, id uuid.UUID) (*Session, error) {
	s := &Session{ID: id}
	err := db.QueryRowContext(ctx,
		`SELECT user_uuid, game_id, started_at, used
		 FROM sessions WHERE id = $1`, id,
	).Scan(&s.UserUUID, &s.GameID, &s.StartedAt, &s.Used)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

// MarkSessionUsed sets a session as used so it cannot be reused.
func MarkSessionUsed(ctx context.Context, db *sql.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx,
		`UPDATE sessions SET used = TRUE WHERE id = $1 AND used = FALSE`, id,
	)
	if err != nil {
		return fmt.Errorf("mark session used: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("session already used or not found")
	}
	return nil
}
