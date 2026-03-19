package validate

import (
	"fmt"
	"regexp"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

// Score checks that a score value is within acceptable bounds.
func Score(score, maxScore int64) error {
	if score < 0 {
		return fmt.Errorf("score must not be negative")
	}
	if score > maxScore {
		return fmt.Errorf("score exceeds maximum allowed value (%d)", maxScore)
	}
	return nil
}

// Username checks that a username meets format requirements.
func Username(name string) error {
	if !usernamePattern.MatchString(name) {
		return fmt.Errorf("username must be 3-32 characters, alphanumeric, hyphens, or underscores")
	}
	return nil
}

// GameSlug checks that a game slug is non-empty and reasonable.
func GameSlug(slug string) error {
	if len(slug) == 0 || len(slug) > 64 {
		return fmt.Errorf("game slug must be 1-64 characters")
	}
	return nil
}
