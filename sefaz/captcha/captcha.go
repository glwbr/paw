// Package captcha defines the shared interface and types for captcha solving.
package captcha

import "context"

// Challenge holds an image-based captcha challenge from a portal.
type Challenge struct {
	ID          string // unique identifier (e.g., timestamp used in the request)
	Image       []byte // raw image bytes (JPEG or PNG)
	ContentType string // MIME type of the image
}

// Solution is the resolved answer to a Challenge.
type Solution struct {
	Text        string // the captcha text entered or computed
	ChallengeID string // ties back to the originating Challenge
}

// Solver resolves captcha challenges. Implementations may be manual
// (terminal prompt), automated (OCR service), or remote (API call).
type Solver interface {
	Solve(ctx context.Context, challenge Challenge) (Solution, error)
}
