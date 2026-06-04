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

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return nil, context.Canceled
}

func TestCreateImport_Validation(t *testing.T) {
	srv := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &dummyFetcher{} })

	cases := []struct {
		name   string
		body   map[string]string
		status int
	}{
		{
			name:   "invalid access key format",
			body:   map[string]string{"access_key": "123"},
			status: http.StatusBadRequest,
		},
		{
			name:   "missing both",
			body:   map[string]string{},
			status: http.StatusBadRequest,
		},
		{
			name:   "both provided",
			body:   map[string]string{"access_key": "44000000000000000000000000000000000000000000", "qr_url": "http://example.com"},
			status: http.StatusBadRequest,
		},
		{
			name:   "valid access key",
			body:   map[string]string{"access_key": "44000000000000000000000000000000000000000000"},
			status: http.StatusAccepted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != tc.status {
				t.Errorf("got status %d, want %d", w.Code, tc.status)
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("44000000000000000000000000000000000000000000", "")

	// Test without captcha
	b, _ := json.Marshal(ri)
	var res map[string]any
	json.Unmarshal(b, &res)
	if _, ok := res["captcha_url"]; ok {
		t.Error("captcha_url should not be present when status is pending")
	}

	// Set status to waiting_captcha
	ri.setStatus(StatusWaitingCaptcha)
	b, _ = json.Marshal(ri)
	json.Unmarshal(b, &res)

	wantURL := "/receipts/imports/" + ri.ID + "/captcha"
	if got := res["captcha_url"]; got != wantURL {
		t.Errorf("got captcha_url %q, want %q", got, wantURL)
	}
}
