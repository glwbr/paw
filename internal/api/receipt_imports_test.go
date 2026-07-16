package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

	t.Run("invalid access key", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"access_key": "123"})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := s.createImport(w, req)
		if err == nil {
			t.Fatal("expected error for invalid access key, got nil")
		}

		if msg := err.Error(); msg == "" {
			t.Errorf("expected error message, got empty")
		}
	})

	t.Run("valid access key", func(t *testing.T) {
		validKey := "12345678901234567890123456789012345678901234" // 44 digits
		body, _ := json.Marshal(map[string]string{"access_key": validKey})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		w := httptest.NewRecorder()

		err := s.createImport(w, req)
		if err != nil {
			t.Fatalf("expected no error for valid access key, got %v", err)
		}
		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}
	})
}
