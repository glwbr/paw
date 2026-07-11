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
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if res["captcha_url"] != "/receipts/imports/test-id/captcha" {
		t.Errorf("expected captcha_url /receipts/imports/test-id/captcha, got %v", res["captcha_url"])
	}

	ri.status = StatusCompleted
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("expected no captcha_url when status is not waiting_captcha, got %v", res["captcha_url"])
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &mockFetcher{} })

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid access key length",
			body:           map[string]string{"access_key": "123"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "access key must be 44 digits",
		},
		{
			name:           "invalid access key characters",
			body:           map[string]string{"access_key": "1234567890123456789012345678901234567890123a"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "access key must be 44 numeric digits",
		},
		{
			name:           "valid access key",
			body:           map[string]string{"access_key": "12345678901234567890123456789012345678901234"},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			w := httptest.NewRecorder()

			err := s.createImport(w, req)

			if tt.expectedStatus == http.StatusBadRequest {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var he *apierrors.HTTPError
				if !errors.As(err, &he) {
					t.Fatalf("expected HTTPError, got %T", err)
				}
				if he.Status != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, he.Status)
				}
				if he.Msg != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, he.Msg)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if w.Code != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
				}
			}
		})
	}
}
