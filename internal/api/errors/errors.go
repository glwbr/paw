// Package errors provides HTTP-aware error types for the API layer.
// It is app-private (under internal/) because HTTP status codes are an
// application concern; library packages (sefaz/, nfce/) must not produce
// these errors.
//
// Import with an alias to avoid shadowing the stdlib errors package:
//
//	import apierrors "github.com/glwbr/paw/internal/api/errors"
package errors

import (
	"fmt"
	"net/http"

	"github.com/glwbr/paw/errs"
)

// HTTPError carries a safe-to-expose message and an HTTP status code.
// It implements errs.Public for central writeError dispatch.
type HTTPError struct {
	Status int
	Msg    string // safe to expose to clients
	Err    error  // internal cause; logged, never serialized
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %s", e.Status, e.Msg, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Msg)
}

func (e *HTTPError) PublicMessage() string { return e.Msg }
func (e *HTTPError) Unwrap() error         { return e.Err }

// Compile-time check: HTTPError satisfies errs.Public.
var _ errs.Public = (*HTTPError)(nil)

func BadRequest(msg string, cause error) *HTTPError {
	return &HTTPError{Status: http.StatusBadRequest, Msg: msg, Err: cause}
}

func NotFound(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusNotFound, Msg: msg}
}

func Conflict(msg string) *HTTPError {
	return &HTTPError{Status: http.StatusConflict, Msg: msg}
}

// Internal wraps an opaque error as a 500; client receives only "internal error" (cause logged server-side).
func Internal(cause error) *HTTPError {
	return &HTTPError{Status: http.StatusInternalServerError, Msg: "internal error", Err: cause}
}
