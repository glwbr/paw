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
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type mockFetcher struct{}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{}, nil
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expected := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != expected {
		t.Errorf("expected captcha_url %q, got %q", expected, res["captcha_url"])
	}

	ri.status = StatusPending
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted when status is not waiting_captcha")
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	t.Run("invalid access key length", func(t *testing.T) {
		body, err := json.Marshal(map[string]string{"access_key": "123"})
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err = s.createImport(w, req)
		if err == nil {
			t.Fatal("expected error for invalid access key, got nil")
		}

		var httpErr *apierrors.HTTPError
		if !errors.As(err, &httpErr) {
			t.Fatalf("expected HTTPError, got %T: %v", err, err)
		}

		if httpErr.Status != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, httpErr.Status)
		}

		expectedMsg := "access key must be 44 numeric digits"
		if httpErr.Msg != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, httpErr.Msg)
		}
	})

	t.Run("valid access key", func(t *testing.T) {
		validKey := "29240112345678000101650010000000011000000019" // 44 digits
		body, err := json.Marshal(map[string]string{"access_key": validKey})
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err = s.createImport(w, req)
		if err != nil {
			t.Fatalf("expected no error for valid access key, got %v", err)
		}
		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}
	})
}
