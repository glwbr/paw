package nfce

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/glwbr/paw/nfce/internal/parse"
)

// MeasureUnit represents units of measurement (KG, L, G, ML, M).
type MeasureUnit string

// PackUnit represents units of quantity/packing (PC, CX, PCT, DZ).
type PackUnit string

// UnitKind indicates the type of unit (measure or pack).
type UnitKind int

const (
	UnitKindUnknown UnitKind = iota
	UnitKindMeasure
	UnitKindPack
)

// NormalizedUnit represents a classified unit, either measure-based or pack-based.
type NormalizedUnit struct {
	Kind        UnitKind
	MeasureUnit MeasureUnit
	PackUnit    PackUnit
}

const (
	MeasureKg         MeasureUnit = "KG"
	MeasureGram       MeasureUnit = "G"
	MeasureLiter      MeasureUnit = "L"
	MeasureMilliliter MeasureUnit = "ML"
	MeasureMeter      MeasureUnit = "M"
)

const (
	PackUnitPiece PackUnit = "PC"  // Unit / piece
	PackUnitBox   PackUnit = "CX"  // Box
	PackUnitPack  PackUnit = "PCT" // Pack
	PackUnitDozen PackUnit = "DZ"  // Dozen
)

func NormalizeUnit(raw string) NormalizedUnit {
	if strings.TrimSpace(raw) == "" {
		return NormalizedUnit{Kind: UnitKindUnknown}
	}

	n := parse.Normalize(raw)

	switch {
	case parse.ContainsAny(n, "KG", "KILO", "QUILO"):
		return NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureKg}

	case parse.ContainsAny(n, "ML", "MILILITRO"):
		return NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureMilliliter}

	case parse.ContainsAny(n, "L", "LT", "LITRO"):
		return NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureLiter}

	case parse.ContainsAny(n, "G", "GR", "GRAMA"):
		return NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureGram}

	case parse.ContainsAny(n, "M", "MT", "METRO"):
		return NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureMeter}

	case parse.ContainsAny(n, "PC", "UN", "UND", "UNID", "UNIDADE"):
		return NormalizedUnit{Kind: UnitKindPack, PackUnit: PackUnitPiece}

	case parse.ContainsAny(n, "CX", "CAIXA"):
		return NormalizedUnit{Kind: UnitKindPack, PackUnit: PackUnitBox}

	case parse.ContainsAny(n, "PCT", "PACOTE", "PAC", "PACK", "PK"):
		return NormalizedUnit{Kind: UnitKindPack, PackUnit: PackUnitPack}

	case parse.ContainsAny(n, "DZ", "DUZIA"):
		return NormalizedUnit{Kind: UnitKindPack, PackUnit: PackUnitDozen}

	default:
		return NormalizedUnit{Kind: UnitKindUnknown}
	}
}

func (u MeasureUnit) BaseUnit() MeasureUnit {
	switch u {
	case MeasureGram:
		return MeasureKg
	case MeasureMilliliter:
		return MeasureLiter
	default:
		return u
	}
}

// ConversionRatio returns the numerator and denominator for converting
// this unit to its base unit (e.g. G→KG returns 1, 1000).
func (u MeasureUnit) ConversionRatio() (num, den int64) {
	switch u {
	case MeasureGram:
		return 1, 1000
	case MeasureMilliliter:
		return 1, 1000
	default:
		return 1, 1
	}
}

func (u MeasureUnit) IsWeight() bool { return u == MeasureKg || u == MeasureGram }

func (u MeasureUnit) IsVolume() bool { return u == MeasureLiter || u == MeasureMilliliter }

// PricePerBaseUnit calculates normalized price (e.g. price per KG or per L).
func PricePerBaseUnit(price Money, qty Quantity, u NormalizedUnit) (Money, error) {
	if u.Kind != UnitKindMeasure {
		return 0, fmt.Errorf("cannot compute price per unit for non-measure unit")
	}
	if qty <= 0 {
		return 0, fmt.Errorf("invalid quantity")
	}
	num, den := u.MeasureUnit.ConversionRatio()
	return Money(int64(price) * QuantityScale * den / (int64(qty) * num)), nil
}

func (u NormalizedUnit) String() string {
	switch u.Kind {
	case UnitKindMeasure:
		return string(u.MeasureUnit)
	case UnitKindPack:
		return string(u.PackUnit)
	default:
		return "UNKNOWN"
	}
}

// Value implements driver.Valuer. Returns the canonical unit code (KG, L, PC, UNKNOWN, etc.).
func (u NormalizedUnit) Value() (driver.Value, error) { return u.String(), nil }

// Scan implements sql.Scanner. Expects a canonical code from the database.
func (u *NormalizedUnit) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case nil:
		*u = NormalizedUnit{Kind: UnitKindUnknown}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into NormalizedUnit", src)
	}
	switch MeasureUnit(s) {
	case MeasureKg, MeasureGram, MeasureLiter, MeasureMilliliter, MeasureMeter:
		*u = NormalizedUnit{Kind: UnitKindMeasure, MeasureUnit: MeasureUnit(s)}
		return nil
	}
	switch PackUnit(s) {
	case PackUnitPiece, PackUnitBox, PackUnitPack, PackUnitDozen:
		*u = NormalizedUnit{Kind: UnitKindPack, PackUnit: PackUnit(s)}
		return nil
	}
	*u = NormalizedUnit{Kind: UnitKindUnknown}
	return nil
}
