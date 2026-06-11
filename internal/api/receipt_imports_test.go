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

func (m *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(c captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid access key",
			body:       map[string]string{"access_key": "12345678901234567890123456789012345678901234"},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid access key length",
			body:       map[string]string{"access_key": "1234"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid access key characters",
			body:       map[string]string{"access_key": "1234567890123456789012345678901234567890123a"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing both",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "both present",
			body:       map[string]string{"access_key": "12345678901234567890123456789012345678901234", "qr_url": "http://example.com"},
			wantStatus: http.StatusBadRequest,
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
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]any
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}

	want := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != want {
		t.Errorf("got captcha_url %v, want %v", res["captcha_url"], want)
	}

	ri.status = StatusCompleted
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted when not waiting for captcha, got %q. JSON: %s", res["captcha_url"], string(b))
	}
}
