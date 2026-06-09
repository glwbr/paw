package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func dummyFetcherFactory(captcha.Solver) sefaz.Fetcher {
	return &dummyFetcher{}
}

func TestCreateImport(t *testing.T) {
	s := NewServer(nil, dummyFetcherFactory)

	tests := []struct {
		name       string
		payload    string
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid access key",
			payload:    `{"access_key": "12345678901234567890123456789012345678901234"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid access key (too short)",
			payload:    `{"access_key": "1234"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "must be 44 numeric digits",
		},
		{
			name:       "invalid access key (non-numeric)",
			payload:    `{"access_key": "1234567890123456789012345678901234567890123a"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "must be 44 numeric digits",
		},
		{
			name:       "missing both",
			payload:    `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "exactly one of access_key or qr_url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tt.payload))
			rr := httptest.NewRecorder()
			s.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantError != "" {
				var resp map[string]string
				_ = json.Unmarshal(rr.Body.Bytes(), &resp)
				if !strings.Contains(resp["error"], tt.wantError) {
					t.Errorf("got error %q, want it to contain %q", resp["error"], tt.wantError)
				}
			}
		})
	}
}

func TestCaptchaURL(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")

	// Case 1: Status Pending (No captcha URL)
	b, _ := json.Marshal(ri)
	var resp map[string]any
	_ = json.Unmarshal(b, &resp)
	if _, ok := resp["captcha_url"]; ok {
		t.Error("captcha_url should not be present in pending status")
	}

	// Case 2: Status Waiting Captcha (Captcha URL present)
	ri.setStatus(StatusWaitingCaptcha)
	b, _ = json.Marshal(ri)
	_ = json.Unmarshal(b, &resp)
	got, ok := resp["captcha_url"].(string)
	if !ok {
		t.Fatal("captcha_url should be present in waiting_captcha status")
	}
	want := "/receipts/imports/" + ri.ID + "/captcha"
	if got != want {
		t.Errorf("got captcha_url %q, want %q", got, want)
	}
}
