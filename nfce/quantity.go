package nfce

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// QuantityScale is the fixed-point scaling factor for Quantity values.
// A raw quantity of 2.5 is stored as Quantity(25000).
const QuantityScale = 10000

// Quantity represents a scaled item quantity (×10000 for precision).
// Example: 2.5 units → Quantity(25000)
type Quantity int64

// Float64 returns the quantity as a float (for display only).
func (q Quantity) Float64() float64 { return float64(q) / QuantityScale }

func (q Quantity) String() string { return fmt.Sprintf("%.4f", q.Float64()) }

// IsZero reports whether q is zero.
func (q Quantity) IsZero() bool { return q == 0 }

// Value implements driver.Valuer (stores as BIGINT).
func (q Quantity) Value() (driver.Value, error) { return int64(q), nil }

// ParseQuantity parses Brazilian decimal strings into Quantity.
// Handles: "0,1880" → 1880, "2,5000" → 25000, "1.000,0000" → 10000000
func ParseQuantity(s string) (Quantity, error) {
	s = parse.Text(s)
	if s == "" {
		return 0, fmt.Errorf("parse quantity: %w", parse.ErrEmptyInput)
	}
	f, err := strconv.ParseFloat(parse.BRNumber(s), 64)
	if err != nil {
		return 0, fmt.Errorf("parse quantity %q: %w", s, err)
	}
	return Quantity(math.Round(f * QuantityScale)), nil
}

// Scan implements sql.Scanner (reads BIGINT/INT/SMALLINT).
func (q *Quantity) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*q = Quantity(v)
	case int32:
		*q = Quantity(v)
	case int:
		*q = Quantity(v)
	case nil:
		*q = 0
	default:
		return fmt.Errorf("cannot scan %T into Quantity", src)
	}
	return nil
}
