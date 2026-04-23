// Package sefaz provides fetchers for retrieving raw NFC-e HTML pages
// from Brazilian state SEFAZ portals.
package sefaz

import "context"

// Fetcher retrieves raw HTML pages from a state's SEFAZ NFC-e portal.
type Fetcher interface {
	Fetch(ctx context.Context, req *Request) (*Result, error)
}

// Request specifies what to fetch. Exactly one of AccessKey or QRURL must be set.
type Request struct {
	// AccessKey is the 44-digit NFC-e chave de acesso.
	// Requires captcha solving when supported by the state fetcher.
	AccessKey string

	// QRURL is the full QR code URL from a physical receipt.
	// Typically bypasses captcha.
	QRURL string
}

// Result holds the raw HTML page fetched from the portal.
type Result struct {
	// Page contains the raw HTML of the print page.
	Page []byte
}
