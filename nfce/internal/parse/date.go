package parse

import (
	"errors"
	"fmt"
	"time"
)

// ErrEmptyInput is returned when a parsing function receives an empty string.
var ErrEmptyInput = errors.New("empty input")

// Date parses common Brazilian date formats.
func Date(s string) (time.Time, error) {
	s = Text(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("parse date: %w", ErrEmptyInput)
	}

	loc := time.Local

	layouts := []string{
		"02/01/2006 15:04:05-07:00",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006",
		time.RFC3339,
	}

	var lastErr error

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, loc)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("parse date %q: %w", s, lastErr)
}
