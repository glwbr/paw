package nfce

import (
	"database/sql/driver"
	"fmt"
)

// UF represents a Brazilian state code.
type UF string

const (
	BA UF = "BA"
	// MG is unsupported; present for testing only.
	MG UF = "MG"
)

var ibgeCodeMap = map[int]UF{
	11: "RO", 12: "AC", 13: "AM", 14: "RR", 15: "PA", 16: "AP", 17: "TO",
	21: "MA", 22: "PI", 23: "CE", 24: "RN", 25: "PB", 26: "PE", 27: "AL", 28: "SE", 29: "BA",
	31: "MG", 32: "ES", 33: "RJ", 35: "SP",
	41: "PR", 42: "SC", 43: "RS",
	50: "MS", 51: "MT", 52: "GO", 53: "DF",
}

// UFFromCode returns the UF for a given IBGE numeric state code.
func UFFromCode(code int) (UF, bool) {
	uf, ok := ibgeCodeMap[code]
	return uf, ok
}

func (u UF) String() string { return string(u) }

// Value implements driver.Valuer.
func (u UF) Value() (driver.Value, error) { return string(u), nil }

// Scan implements sql.Scanner.
func (u *UF) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*u = UF(v)
		return nil
	case []byte:
		*u = UF(v)
		return nil
	case nil:
		*u = ""
		return nil
	default:
		return fmt.Errorf("cannot scan %T into UF", src)
	}
}
