package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.ID = "test-id"

	// Initial status: pending
	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	var res map[string]any
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Error("expected no captcha_url when status is pending")
	}

	// Change status to waiting_captcha
	ri.status = StatusWaitingCaptcha
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}
	expectedURL := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %q", expectedURL, res["captcha_url"])
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return nil })

	t.Run("invalid access key length", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"access_key": "123"})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := s.createImport(w, req)
		var he *apierrors.HTTPError
		if !errors.As(err, &he) {
			t.Fatalf("expected HTTPError, got %T", err)
		}
		if he.Status != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", he.Status)
		}
		if he.Msg != "access key must be 44 numeric digits" {
			t.Errorf("expected message %q, got %q", "access key must be 44 numeric digits", he.Msg)
		}
	})

	t.Run("invalid access key characters", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"access_key": "1234567890123456789012345678901234567890123a"})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := s.createImport(w, req)
		var he *apierrors.HTTPError
		if !errors.As(err, &he) {
			t.Fatalf("expected HTTPError, got %T", err)
		}
		if he.Msg != "access key must be 44 numeric digits" {
			t.Errorf("expected message %q, got %q", "access key must be 44 numeric digits", he.Msg)
		}
	})
}
