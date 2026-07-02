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

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid access key",
			body:       `{"access_key": "29240112345678901234650010000000011234567890"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid access key (short)",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid access_key",
		},
		{
			name:       "invalid access key (non-numeric)",
			body:       `{"access_key": "2924011234567890123465001000000001123456789A"}`,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			s.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantError != "" {
				var res map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
					t.Fatalf("failed to unmarshal error body: %v", err)
				}
				if !strings.Contains(res["error"], tt.wantError) {
					t.Errorf("got error %q, want it to contain %q", res["error"], tt.wantError)
				}
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("29240112345678901234650010000000011234567890", "")

	// Case 1: Status Pending (No captcha_url)
	ri.setStatus(StatusPending)
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted for status %s", ri.Status())
	}

	// Case 2: Status WaitingCaptcha (With captcha_url)
	ri.setStatus(StatusWaitingCaptcha)
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	res = make(map[string]any) // Reset map
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	wantURL := "/receipts/imports/" + ri.ID + "/captcha"
	if res["captcha_url"] != wantURL {
		t.Errorf("got captcha_url %q, want %q", res["captcha_url"], wantURL)
	}
}
