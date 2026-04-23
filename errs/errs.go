// Package errs defines the Public interface that errors implement to declare
// their message is safe to surface to end users
package errs

import "errors"

// Public is implemented by errors whose message is safe to surface to
// API clients. Anything that does not implement this interface is treated
// as internal and replaced with a generic message at every service boundary.
type Public interface {
	error
	PublicMessage() string
}

// Sentinel is a package-level error whose message is safe to expose to
// clients. Error() and PublicMessage() both return Msg.
//
// Use for sentinels where the message is the whole signal (no runtime
// data to carry). When you need to attach fields (access key, state, etc.)
// define a dedicated struct type with its own PublicMessage() method.
//
//	var ErrMissingInput = &errs.Sentinel{Msg: "access key or QR URL is required"}
type Sentinel struct{ Msg string }

func (e *Sentinel) Error() string         { return e.Msg }
func (e *Sentinel) PublicMessage() string { return e.Msg }

// PublicMessage walks the error chain via errors.As and returns the first
// Public message found, or "" if no error in the chain is Public.
func PublicMessage(err error) string {
	var p Public
	if errors.As(err, &p) {
		return p.PublicMessage()
	}
	return ""
}
