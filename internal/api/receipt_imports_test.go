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

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	tests := []struct {
		name        string
		status      Status
		wantCaptcha bool
	}{
		{
			name:        "pending status - no captcha url",
			status:      StatusPending,
			wantCaptcha: false,
		},
		{
			name:        "waiting_captcha status - has captcha url",
			status:      StatusWaitingCaptcha,
			wantCaptcha: true,
		},
		{
			name:        "completed status - no captcha url",
			status:      StatusCompleted,
			wantCaptcha: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri := newReceiptImport("44240112345678901234650010000000011234567890", "")
			ri.status = tt.status

			data, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("failed to marshal ReceiptImport: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(data, &res); err != nil {
				t.Fatalf("failed to unmarshal ReceiptImport JSON: %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tt.wantCaptcha {
				if !ok || url == "" {
					t.Errorf("expected captcha_url to be present, got %v", res["captcha_url"])
				}
				wantURL := "/receipts/imports/" + ri.ID + "/captcha"
				if url != wantURL {
					t.Errorf("want captcha_url %q, got %q", wantURL, url)
				}
			} else {
				if ok {
					t.Errorf("expected captcha_url to be omitted, got %q", url)
				}
			}
		})
	}
}

type mockFetcher struct{}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{}, nil
}

func TestServer_CreateImport_Validation(t *testing.T) {
	// dummyFetcher implements sefaz.Fetcher to satisfy the Server's requirement.
	// We provide a no-op fetcher factory to prevent panics in the background goroutine.
	dummyFetcher := func(captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	}
	s := NewServer(nil, dummyFetcher)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantErrMsg string
	}{
		{
			name: "valid access key",
			body: map[string]string{
				"access_key": "44240112345678901234650010000000011234567890",
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "invalid access key format",
			body: map[string]string{
				"access_key": "123",
			},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid access key format",
		},
		{
			name: "missing both access_key and qr_url",
			body: map[string]string{},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "exactly one of access_key or qr_url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)
			if err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantErrMsg != "" {
				var res map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if res["error"] != tt.wantErrMsg {
					t.Errorf("want error message %q, got %q", tt.wantErrMsg, res["error"])
				}
			}
		})
	}
}
