package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/internal/db"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

// DummyFetcher implements sefaz.Fetcher to satisfy the Server's requirement
// and avoid nil pointer dereferences during async imports.
type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, r *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	t.Run("omits captcha_url when status is not waiting_captcha", func(t *testing.T) {
		ri := newReceiptImport("29240112345678901234650010000000012345678901", "")
		ri.status = StatusPending

		data, err := ri.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		res := make(map[string]any)
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatalf("unexpected error unmarshaling: %v", err)
		}

		if _, exists := res["captcha_url"]; exists {
			t.Errorf("expected captcha_url to be omitted, but got: %v", res["captcha_url"])
		}
	})

	t.Run("includes captcha_url when status is waiting_captcha", func(t *testing.T) {
		ri := newReceiptImport("29240112345678901234650010000000012345678901", "")
		ri.status = StatusWaitingCaptcha

		data, err := ri.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error marshaling: %v", err)
		}

		res := make(map[string]any)
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatalf("unexpected error unmarshaling: %v", err)
		}

		expected := "/receipts/imports/" + ri.ID + "/captcha"
		if got, exists := res["captcha_url"]; !exists || got != expected {
			t.Errorf("expected captcha_url to be %q, got: %v", expected, got)
		}
	})
}

func TestCreateImport_Validation(t *testing.T) {
	q := &db.Queries{}
	server := NewServer(q, func(solver captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	t.Run("returns bad request for invalid access key length", func(t *testing.T) {
		body := []byte(`{"access_key": "123"}`)
		req := httptest.NewRequest(http.MethodPost, "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := server.createImport(w, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var he *apierrors.HTTPError
		if !errors.As(err, &he) {
			t.Fatalf("expected HTTPError, got %T: %v", err, err)
		}

		if he.Status != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", he.Status)
		}

		if he.Msg != "access key must be 44 numeric digits" {
			t.Errorf("expected public message %q, got %q", "access key must be 44 numeric digits", he.Msg)
		}
	})

	t.Run("returns bad request for non-numeric access key", func(t *testing.T) {
		body := []byte(`{"access_key": "292401123456789012346500100000000123456789ab"}`)
		req := httptest.NewRequest(http.MethodPost, "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := server.createImport(w, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var he *apierrors.HTTPError
		if !errors.As(err, &he) {
			t.Fatalf("expected HTTPError, got %T: %v", err, err)
		}

		if he.Status != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", he.Status)
		}

		if he.Msg != "access key must be 44 numeric digits" {
			t.Errorf("expected public message %q, got %q", "access key must be 44 numeric digits", he.Msg)
		}
	})

	t.Run("accepts valid access key", func(t *testing.T) {
		body := []byte(`{"access_key": "29240112345678901234650010000000012345678901"}`)
		req := httptest.NewRequest(http.MethodPost, "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := server.createImport(w, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}
	})
}
