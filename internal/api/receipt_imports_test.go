package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return nil, errors.New("not implemented")
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &dummyFetcher{} })

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid access key",
			body:       map[string]string{"access_key": "29240112345678000190650010000034026311019296"},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid access key - too short",
			body:       map[string]string{"access_key": "12345"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid access key - too long",
			body:       map[string]string{"access_key": "29240112345678000190650010000034026311019296123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid access key - non-digits",
			body:       map[string]string{"access_key": "2924011234567800019065001000003402631101929A"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid qr url",
			body:       map[string]string{"qr_url": "https://example.com/nfce"},
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)
			if err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d. Body: %s", rr.Code, tc.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("29240112345678000190650010000034026311019296", "")
	ri.ID = "test-id"

	tests := []struct {
		name    string
		status  Status
		wantURL string
	}{
		{
			name:    "pending - no captcha url",
			status:  StatusPending,
			wantURL: "",
		},
		{
			name:    "waiting_captcha - has captcha url",
			status:  StatusWaitingCaptcha,
			wantURL: "/receipts/imports/test-id/captcha",
		},
		{
			name:    "completed - no captcha url",
			status:  StatusCompleted,
			wantURL: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ri.setStatus(tc.status)
			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var res struct {
				CaptchaURL string `json:"captcha_url"`
			}
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if res.CaptchaURL != tc.wantURL {
				t.Errorf("captcha_url = %q, want %q", res.CaptchaURL, tc.wantURL)
			}
		})
	}
}
