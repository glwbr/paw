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
	return &sefaz.Result{}, nil
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	cases := []struct {
		name       string
		req        map[string]string
		wantStatus int
	}{
		{
			name: "invalid access key (too short)",
			req: map[string]string{
				"access_key": "123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid access key (non-numeric)",
			req: map[string]string{
				"access_key": "1234567890123456789012345678901234567890123a",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access key",
			req: map[string]string{
				"access_key": "29240112345678000195650010000000011000000015",
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "missing both",
			req:  map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.req)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			s.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("29240112345678000195650010000000011000000015", "")

	// Pending - no captcha_url
	b, _ := ri.MarshalJSON()
	var res map[string]any
	_ = json.Unmarshal(b, &res)
	if _, ok := res["captcha_url"]; ok {
		t.Error("captcha_url should not be present in pending status")
	}

	// WaitingCaptcha - should have captcha_url
	ri.setStatus(StatusWaitingCaptcha)
	b, _ = ri.MarshalJSON()
	res = make(map[string]any)
	_ = json.Unmarshal(b, &res)

	wantURL := "/receipts/imports/" + ri.ID + "/captcha"
	if got := res["captcha_url"]; got != wantURL {
		t.Errorf("captcha_url: want %q, got %q", wantURL, got)
	}
}
