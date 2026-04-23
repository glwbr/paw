package errors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/glwbr/paw/errs"
	apierrors "github.com/glwbr/paw/internal/api/errors"
)

func TestConstructorStatuses(t *testing.T) {
	cause := errors.New("cause")
	cases := []struct {
		err    *apierrors.HTTPError
		status int
		msg    string
	}{
		{apierrors.BadRequest("bad input", cause), http.StatusBadRequest, "bad input"},
		{apierrors.NotFound("not found"), http.StatusNotFound, "not found"},
		{apierrors.Conflict("conflict"), http.StatusConflict, "conflict"},
		{apierrors.Internal(cause), http.StatusInternalServerError, "internal error"},
	}
	for _, tc := range cases {
		if tc.err.Status != tc.status {
			t.Errorf("%T: want status %d, got %d", tc.err, tc.status, tc.err.Status)
		}
		if tc.err.PublicMessage() != tc.msg {
			t.Errorf("%T: want msg %q, got %q", tc.err, tc.msg, tc.err.PublicMessage())
		}
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := apierrors.BadRequest("bad", cause)
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should find root cause through Unwrap")
	}
}

func TestSatisfiesErrsPublic(t *testing.T) {
	var _ errs.Public = (*apierrors.HTTPError)(nil)
	err := apierrors.NotFound("missing")
	if got := errs.PublicMessage(err); got != "missing" {
		t.Fatalf("errs.PublicMessage: want %q, got %q", "missing", got)
	}
}

func TestError_withCause(t *testing.T) {
	err := apierrors.BadRequest("bad", errors.New("detail"))
	if err.Error() == "" {
		t.Fatal("Error() should not be empty")
	}
}

func TestError_withoutCause(t *testing.T) {
	err := apierrors.NotFound("gone")
	if err.Unwrap() != nil {
		t.Fatal("Unwrap should be nil when no cause provided")
	}
}
