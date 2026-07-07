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

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")

	// Test default state
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Error("expected no captcha_url in pending status")
	}

	// Test waiting_captcha state
	ri.setStatus(StatusWaitingCaptcha)
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}
	gotURL, ok := res["captcha_url"].(string)
	if !ok {
		t.Fatal("expected captcha_url in waiting_captcha status")
	}
	wantURL := "/receipts/imports/" + ri.ID + "/captcha"
	if gotURL != wantURL {
		t.Errorf("want captcha_url %q, got %q", wantURL, gotURL)
	}
}

type mockFetcher struct{}
func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{}, nil
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantError  string
	}{
		{
			name: "invalid access key format",
			body: map[string]string{
				"access_key": "invalid",
			},
			wantError:  "access key must be 44 numeric digits",
		},
		{
			name: "valid access key format",
			body: map[string]string{
				"access_key": "12345678901234567890123456789012345678901234",
			},
			wantError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)

			if tt.wantError != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var he *apierrors.HTTPError
				if !errors.As(err, &he) {
					t.Fatalf("expected *apierrors.HTTPError, got %T", err)
				}
				if he.Msg != tt.wantError {
					t.Errorf("want error msg %q, got %q", tt.wantError, he.Msg)
				}
				if he.Status != http.StatusBadRequest {
					t.Errorf("want status %d, got %d", http.StatusBadRequest, he.Status)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if rr.Code != http.StatusAccepted {
					t.Errorf("want status %d, got %d", http.StatusAccepted, rr.Code)
				}
			}
		})
	}
}
