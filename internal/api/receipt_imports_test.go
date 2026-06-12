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

func newDummyFetcher(captcha.Solver) sefaz.Fetcher {
	return &dummyFetcher{}
}

func TestCreateImport_Validation(t *testing.T) {
	srv := NewServer(nil, newDummyFetcher)

	cases := []struct {
		name string
		body map[string]string
		want int
	}{
		{
			name: "Valid Access Key",
			body: map[string]string{"access_key": "12345678901234567890123456789012345678901234"},
			want: http.StatusAccepted,
		},
		{
			name: "Invalid Access Key (too short)",
			body: map[string]string{"access_key": "123"},
			want: http.StatusBadRequest,
		},
		{
			name: "Invalid Access Key (non-numeric)",
			body: map[string]string{"access_key": "notnumeric1234567890123456789012345678901234"},
			want: http.StatusBadRequest,
		},
		{
			name: "Missing both",
			body: map[string]string{},
			want: http.StatusBadRequest,
		},
		{
			name: "Both provided",
			body: map[string]string{"access_key": "12345678901234567890123456789012345678901234", "qr_url": "http://example.com"},
			want: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			if err := srv.createImport(rr, req); err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tc.want {
				t.Errorf("got status %d, want %d", rr.Code, tc.want)
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
	if got := res["captcha_url"]; got != want {
		t.Errorf("captcha_url: got %q, want %q", got, want)
	}

	ri.status = StatusCompleted
	b, _ = json.Marshal(ri)
	res = make(map[string]any)
	_ = json.Unmarshal(b, &res)
	if _, ok := res["captcha_url"]; ok {
		t.Error("captcha_url should be omitted when not waiting_captcha")
	}
}
