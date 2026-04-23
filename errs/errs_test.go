package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/glwbr/paw/errs"
)

type publicErr struct{ msg string }

func (e *publicErr) Error() string         { return e.msg }
func (e *publicErr) PublicMessage() string { return e.msg }

type opaqueErr struct{ msg string }

func (e *opaqueErr) Error() string { return e.msg }

func TestPublicMessage_direct(t *testing.T) {
	err := &publicErr{"safe message"}
	if got := errs.PublicMessage(err); got != "safe message" {
		t.Fatalf("want %q, got %q", "safe message", got)
	}
}

func TestPublicMessage_wrapped(t *testing.T) {
	inner := &publicErr{"deep message"}
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("mid: %w", inner))
	if got := errs.PublicMessage(wrapped); got != "deep message" {
		t.Fatalf("want %q, got %q", "deep message", got)
	}
}

func TestPublicMessage_opaque(t *testing.T) {
	err := &opaqueErr{"internal details"}
	if got := errs.PublicMessage(err); got != "" {
		t.Fatalf("want empty string, got %q", got)
	}
}

func TestPublicMessage_nil(t *testing.T) {
	if got := errs.PublicMessage(nil); got != "" {
		t.Fatalf("want empty string for nil, got %q", got)
	}
}

func TestPublicMessage_publicOuterOpaqueInner(t *testing.T) {
	inner := &opaqueErr{"internal"}
	outer := &publicErr{fmt.Sprintf("safe: %s", inner)}
	wrapped := fmt.Errorf("wrap: %w", outer)
	if got := errs.PublicMessage(wrapped); got != outer.PublicMessage() {
		t.Fatalf("want %q, got %q", outer.PublicMessage(), got)
	}
}

func TestPublicInterface_satisfaction(t *testing.T) {
	var _ errs.Public = &publicErr{}
	// Verify errors.As works through the chain.
	err := fmt.Errorf("a: %w", &publicErr{"msg"})
	var p errs.Public
	if !errors.As(err, &p) {
		t.Fatal("errors.As should find Public in chain")
	}
}
