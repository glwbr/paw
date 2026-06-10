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
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &dummyFetcher{} })

	cases := []struct {
		name   string
		body   string
		status int
		err    string
	}{
		{
			name:   "invalid access key length",
			body:   `{"access_key": "123"}`,
			status: http.StatusBadRequest,
			err:    "access key must be exactly 44 digits",
		},
		{
			name:   "invalid access key characters",
			body:   `{"access_key": "1234567890123456789012345678901234567890123a"}`,
			status: http.StatusBadRequest,
			err:    "access key must be exactly 44 digits",
		},
		{
			name:   "missing both",
			body:   `{}`,
			status: http.StatusBadRequest,
			err:    "exactly one of access_key or qr_url is required",
		},
		{
			name:   "both provided",
			body:   `{"access_key": "12345678901234567890123456789012345678901234", "qr_url": "http://example.com"}`,
			status: http.StatusBadRequest,
			err:    "exactly one of access_key or qr_url is required",
		},
		{
			name:   "valid access key",
			body:   `{"access_key": "12345678901234567890123456789012345678901234"}`,
			status: http.StatusAccepted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			s.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Errorf("status: want %d, got %d", tc.status, rr.Code)
			}
			if tc.err != "" {
				var res map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
					t.Fatal(err)
				}
				if res["error"] != tc.err {
					t.Errorf("error: want %q, got %q", tc.err, res["error"])
				}
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

	var res struct {
		CaptchaURL string `json:"captcha_url"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}

	expected := "/receipts/imports/test-id/captcha"
	if res.CaptchaURL != expected {
		t.Errorf("captcha_url: want %q, got %q", expected, res.CaptchaURL)
	}

	// Verify it's omitted when not waiting_captcha
	ri.status = StatusFetching
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	var res2 map[string]any
	if err := json.Unmarshal(b, &res2); err != nil {
		t.Fatal(err)
	}
	if _, ok := res2["captcha_url"]; ok {
		t.Error("captcha_url should be omitted when status is not waiting_captcha")
	}
}
