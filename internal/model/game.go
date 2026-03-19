package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Game represents a registered game with its own leaderboard.
type Game struct {
	ID        int       `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// GetGameBySlug retrieves a game by its URL-friendly slug.
func GetGameBySlug(ctx context.Context, db *sql.DB, slug string) (*Game, error) {
	g := &Game{Slug: slug}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM games WHERE slug = $1`, slug,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get game by slug: %w", err)
	}
	return g, nil
}

// GetOrCreateGame retrieves a game by slug, creating it if it doesn't exist.
func GetOrCreateGame(ctx context.Context, db *sql.DB, slug string) (*Game, error) {
	g := &Game{Slug: slug}
	err := db.QueryRowContext(ctx,
		`INSERT INTO games (slug, name) VALUES ($1, $1)
		 ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		 RETURNING id, name, created_at`,
		slug,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get or create game: %w", err)
	}
	return g, nil
}
