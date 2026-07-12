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

func TestMarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expected := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != expected {
		t.Errorf("expected captcha_url %q, got %q", expected, res["captcha_url"])
	}

	// Test StatusPending doesn't have captcha_url
	ri.status = StatusPending
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted when status is not waiting_captcha")
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Invalid Access Key Length",
			body: map[string]string{
				"access_key": "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "access key must be 44 numeric digits",
		},
		{
			name: "Invalid Access Key Characters",
			body: map[string]string{
				"access_key": "4444444444444444444444444444444444444444444a",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "access key must be 44 numeric digits",
		},
		{
			name: "Valid Access Key",
			body: map[string]string{
				"access_key": "44444444444444444444444444444444444444444444",
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			// Handle directly since NewServer doesn't expose mux easily without ServeHTTP
			err := s.createImport(rr, req)

			if tt.expectedStatus == http.StatusAccepted {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if rr.Code != http.StatusAccepted {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
				}
			} else {
				var he *apierrors.HTTPError
				if !errors.As(err, &he) {
					t.Fatalf("expected HTTPError, got %T", err)
				}
				if he.Status != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, he.Status)
				}
				if he.Msg != tt.expectedError {
					t.Errorf("expected error message %q, got %q", tt.expectedError, he.Msg)
				}
			}
		})
	}
}
