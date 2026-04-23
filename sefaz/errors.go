package sefaz

import (
	"fmt"
	"time"

	"github.com/glwbr/paw/errs"
)

// Sentinel errors for use with errors.Is.
var (
	ErrMissingInput    = &errs.Sentinel{Msg: "either access key or QR URL must be provided"}
	ErrAmbiguousInput  = &errs.Sentinel{Msg: "provide either access key or QR URL, not both"}
	ErrNoCaptchaSolver = &errs.Sentinel{Msg: "access key lookup requires a captcha solver"}
	ErrCaptchaInvalid  = &errs.Sentinel{Msg: "captcha was rejected by the portal"}
)

// InvalidAccessKeyError is returned when the portal rejects the access key.
type InvalidAccessKeyError struct {
	AccessKey string
}

func (e *InvalidAccessKeyError) Error() string {
	return fmt.Sprintf("invalid access key: %s", e.AccessKey)
}
func (e *InvalidAccessKeyError) PublicMessage() string {
	return "access key was rejected by the portal"
}

// InvoiceNotFoundError is returned when the NFC-e does not exist on the portal.
type InvoiceNotFoundError struct {
	AccessKey string
}

func (e *InvoiceNotFoundError) Error() string {
	return fmt.Sprintf("invoice not found: %s", e.AccessKey)
}
func (e *InvoiceNotFoundError) PublicMessage() string { return "receipt not found on the portal" }

// SessionExpiredError is returned when the portal session times out mid-flow.
type SessionExpiredError struct{}

func (e *SessionExpiredError) Error() string         { return "portal session expired" }
func (e *SessionExpiredError) PublicMessage() string { return "portal session expired" }

// CaptchaError is returned for captcha solver failures.
type CaptchaError struct {
	Err error
}

func (e *CaptchaError) Error() string         { return fmt.Sprintf("captcha error: %s", e.Err) }
func (e *CaptchaError) PublicMessage() string { return "captcha solver failed" }
func (e *CaptchaError) Unwrap() error         { return e.Err }

// RateLimitedError is returned when the portal is rate-limiting requests.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
	}
	return "rate limited by portal"
}
func (e *RateLimitedError) PublicMessage() string { return "portal is rate-limiting requests" }

// PortalError is returned for unexpected portal responses (only HTTP status code is surfaced).
type PortalError struct {
	StatusCode int
	Body       string
}

func (e *PortalError) Error() string {
	return fmt.Sprintf("unexpected portal response (HTTP %d)", e.StatusCode)
}
func (e *PortalError) PublicMessage() string { return e.Error() }
