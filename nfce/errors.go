package nfce

import (
	"fmt"

	"github.com/glwbr/paw/errs"
)

// Sentinel errors for use with errors.Is.
var (
	ErrMissingAccessKey       = &errs.Sentinel{Msg: "missing access key"}
	ErrInvalidAccessKey       = &errs.Sentinel{Msg: "access key must be 44 digits"}
	ErrInvalidAccessKeyFormat = &errs.Sentinel{Msg: "access key must be 44 numeric digits"}
	ErrMissingCNPJ            = &errs.Sentinel{Msg: "missing store CNPJ"}
	ErrNoItems                = &errs.Sentinel{Msg: "no items found"}
	ErrZeroTotal              = &errs.Sentinel{Msg: "total amount is zero"}
	ErrUndetectedState        = &errs.Sentinel{Msg: "could not auto-detect state from HTML content"}
	ErrUnsupportedState       = &errs.Sentinel{Msg: "no parser registered for state"}
)

// ValidationError is returned when a Receipt fails validation checks.
type ValidationError struct {
	Receipt *Receipt
	Err     error
}

func (e *ValidationError) Error() string         { return fmt.Sprintf("validation error: %s", e.Err) }
func (e *ValidationError) PublicMessage() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error         { return e.Err }

// ParseError is returned when a specific field cannot be parsed from raw text.
// It is intentionally not Public — RawText may contain user or portal data.
type ParseError struct {
	Field   string
	RawText string
	Err     error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s from %q: %s", e.Field, e.RawText, e.Err)
}
func (e *ParseError) Unwrap() error { return e.Err }
