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

type mockFetcher struct{}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid access key length",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid access key characters",
			body:       `{"access_key": "4444444444444444444444444444444444444444444X"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid access key (format only)",
			body:       `{"access_key": "44444444444444444444444444444444444444444444"}`,
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			s.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("%s: handler returned wrong status code: got %v want %v", tt.name, rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	data, err := ri.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	wantURL := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != wantURL {
		t.Errorf("got captcha_url %q, want %q", res["captcha_url"], wantURL)
	}

	ri.status = StatusCompleted
	data, err = ri.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted for status %s", ri.status)
	}
}
