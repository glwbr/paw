package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	cases := []struct {
		name    string
		status  Status
		wantURL string
	}{
		{
			name:    "waiting captcha",
			status:  StatusWaitingCaptcha,
			wantURL: "/receipts/imports/123/captcha",
		},
		{
			name:    "pending",
			status:  StatusPending,
			wantURL: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ri := &ReceiptImport{
				ID:     "123",
				status: tc.status,
			}

			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			gotURL, _ := res["captcha_url"].(string)
			if gotURL != tc.wantURL {
				t.Errorf("captcha_url: want %q, got %q", tc.wantURL, gotURL)
			}
		})
	}
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, nil)

	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid access key format",
			body:       `{"access_key": "not-numeric"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid access_key",
		},
		{
			name:       "invalid access key length",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid access_key",
		},
		{
			name:       "missing both",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "exactly one of access_key or qr_url is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var he *apierrors.HTTPError
			if !errors.As(err, &he) {
				t.Fatalf("expected HTTPError, got %T", err)
			}

			if he.Status != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, he.Status)
			}

			if !strings.Contains(he.Msg, tc.wantError) {
				t.Errorf("message: want to contain %q, got %q", tc.wantError, he.Msg)
			}
		})
	}
}
