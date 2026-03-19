package model

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	mathrand "math/rand/v2"
	"time"

	"github.com/google/uuid"
)

// User represents a registered player.
type User struct {
	UUID        uuid.UUID `json:"uuid"`
	Username    string    `json:"username"`
	SecretToken string    `json:"secret_token,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateUser inserts a new user with a randomly generated username and secret token.
// The secret token is returned only on creation and should be persisted by the client.
func CreateUser(ctx context.Context, db *sql.DB, id uuid.UUID) (*User, error) {
	username := generateUsername()
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	u := &User{UUID: id, Username: username, SecretToken: token}

	err = db.QueryRowContext(ctx,
		`INSERT INTO users (uuid, username, secret_token) VALUES ($1, $2, $3)
		 RETURNING created_at, updated_at`,
		id, username, token,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

// GetUser retrieves a user by UUID. Does not include the secret token.
func GetUser(ctx context.Context, db *sql.DB, id uuid.UUID) (*User, error) {
	u := &User{UUID: id}
	err := db.QueryRowContext(ctx,
		`SELECT username, created_at, updated_at FROM users WHERE uuid = $1`, id,
	).Scan(&u.Username, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// ValidateToken checks whether the given token matches the stored token for the user.
func ValidateToken(ctx context.Context, db *sql.DB, id uuid.UUID, token string) (bool, error) {
	var storedToken string
	err := db.QueryRowContext(ctx,
		`SELECT secret_token FROM users WHERE uuid = $1`, id,
	).Scan(&storedToken)
	if err != nil {
		return false, fmt.Errorf("validate token: %w", err)
	}
	return storedToken == token, nil
}

// UpdateUsername changes a user's display name.
func UpdateUsername(ctx context.Context, db *sql.DB, id uuid.UUID, username string) (*User, error) {
	u := &User{UUID: id, Username: username}
	err := db.QueryRowContext(ctx,
		`UPDATE users SET username = $1, updated_at = NOW()
		 WHERE uuid = $2
		 RETURNING created_at, updated_at`,
		username, id,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}
	return u, nil
}

// UserExists checks whether a user with the given UUID exists.
func UserExists(ctx context.Context, db *sql.DB, id uuid.UUID) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE uuid = $1)`, id,
	).Scan(&exists)
	return exists, err
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var adjectives = []string{
	"Swift", "Brave", "Cosmic", "Silent", "Fierce",
	"Lucky", "Mystic", "Rapid", "Golden", "Shadow",
	"Iron", "Crystal", "Thunder", "Frozen", "Blazing",
	"Noble", "Phantom", "Stellar", "Ancient", "Crimson",
}

var nouns = []string{
	"Fox", "Hawk", "Wolf", "Bear", "Dragon",
	"Raven", "Tiger", "Eagle", "Viper", "Falcon",
	"Knight", "Sage", "Ghost", "Phoenix", "Storm",
	"Panther", "Cobra", "Lynx", "Shark", "Titan",
}

func generateUsername() string {
	adj := adjectives[mathrand.IntN(len(adjectives))]
	noun := nouns[mathrand.IntN(len(nouns))]
	num := mathrand.IntN(100)
	return fmt.Sprintf("%s%s%02d", adj, noun, num)
}
