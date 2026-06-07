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
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	cases := []struct {
		name   string
		body   map[string]string
		status int
	}{
		{
			name:   "valid access key",
			body:   map[string]string{"access_key": "12345678901234567890123456789012345678901234"},
			status: http.StatusAccepted,
		},
		{
			name:   "invalid access key length",
			body:   map[string]string{"access_key": "123"},
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid access key chars",
			body:   map[string]string{"access_key": "1234567890123456789012345678901234567890123a"},
			status: http.StatusBadRequest,
		},
		{
			name:   "missing both",
			body:   map[string]string{},
			status: http.StatusBadRequest,
		},
		{
			name:   "both provided",
			body:   map[string]string{"access_key": "12345678901234567890123456789012345678901234", "qr_url": "http://example.com"},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			s.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Errorf("want status %d, got %d. Body: %s", tc.status, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.ID = "test-id"

	// Initial state: no captcha_url
	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := data["captcha_url"]; ok {
		t.Error("captcha_url should NOT be present when status is pending")
	}

	// Waiting captcha state: captcha_url should be present
	ri.setStatus(StatusWaitingCaptcha)
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	want := "/receipts/imports/test-id/captcha"
	if got := data["captcha_url"]; got != want {
		t.Errorf("captcha_url: want %q, got %q", want, got)
	}
}
