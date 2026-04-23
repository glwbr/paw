package nfce

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// Rate represents a tax rate or percentage with 2 decimal places.
// Stored as the value × 100.
// 20.50% = Rate(2050), 7.60% = Rate(760), 1.65% = Rate(165)
type Rate int64

// Percent returns the rate as float64.
func (r Rate) Percent() float64 { return float64(r) / 100 }

func (r Rate) String() string { return fmt.Sprintf("%.2f%%", r.Percent()) }

// ApplyTo returns the tax amount for the given base value.
// Example: Money(1197).ApplyTo(Rate(2050)) = Money(245) (20.50% of R$11.97)
func (r Rate) ApplyTo(base Money) Money {
	return Money((int64(base) * int64(r)) / 10000)
}

// Value implements driver.Valuer.
func (r Rate) Value() (driver.Value, error) { return int64(r), nil }

// ParseRate parses a Brazilian percentage string into a Rate.
// Handles: "20,50" → Rate(2050), "7.60" → Rate(760), "1,65%" → Rate(165)
func ParseRate(s string) (Rate, error) {
	s = parse.Text(s)
	if s == "" {
		return 0, fmt.Errorf("parse rate: %w", parse.ErrEmptyInput)
	}
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	f, err := strconv.ParseFloat(parse.BRNumber(s), 64)
	if err != nil {
		return 0, fmt.Errorf("parse rate %q: %w", s, err)
	}
	return Rate(math.Round(f * 100)), nil
}

// IsZero reports whether r is zero.
func (r Rate) IsZero() bool { return r == 0 }

// Scan implements sql.Scanner (reads BIGINT/INT/SMALLINT).
func (r *Rate) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*r = Rate(v)
	case int32:
		*r = Rate(v)
	case int:
		*r = Rate(v)
	case nil:
		*r = 0
	default:
		return fmt.Errorf("cannot scan %T into Rate", src)
	}
	return nil
}
