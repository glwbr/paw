package nfce

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// Money represents a monetary value in centavos (BRL).
// Example: R$ 12,90 → Money(1290)
type Money int64

// Float64 returns the value in reais (for display only).
func (m Money) Float64() float64 { return float64(m) / 100 }

func (m Money) String() string {
	negative := m < 0
	abs := m
	if negative {
		abs = -abs
	}
	reais := abs / 100
	centavos := abs % 100
	s := fmt.Sprintf("R$ %d,%02d", reais, centavos)
	if negative {
		return "-" + s
	}
	return s
}

func (m Money) Add(other Money) Money { return m + other }

func (m Money) Sub(other Money) Money { return m - other }

// IsZero reports whether m is zero.
func (m Money) IsZero() bool { return m == 0 }

// ParseMoney parses Brazilian currency strings into Money.
// Handles: "1.197,00" → 119700, "12,90" → 1290, "-R$ 10,00" → -1000
func ParseMoney(s string) (Money, error) {
	s = parse.Text(s)
	if s == "" {
		return 0, fmt.Errorf("parse money: %w", parse.ErrEmptyInput)
	}

	s = parse.StripNonNumericPrefix(s)
	negative := strings.Contains(s, "-")
	s = strings.ReplaceAll(s, "-", "")

	f, err := strconv.ParseFloat(parse.BRNumber(s), 64)
	if err != nil {
		return 0, fmt.Errorf("parse money %q: %w", s, err)
	}

	v := Money(math.Round(f * 100))
	if negative {
		v = -v
	}
	return v, nil
}

// Value implements driver.Valuer (stores centavos as BIGINT).
func (m Money) Value() (driver.Value, error) { return int64(m), nil }

// Scan implements sql.Scanner (reads BIGINT/INT/SMALLINT).
func (m *Money) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*m = Money(v)
	case int32:
		*m = Money(v)
	case int:
		*m = Money(v)
	case nil:
		*m = 0
	default:
		return fmt.Errorf("cannot scan %T into Money", src)
	}
	return nil
}
