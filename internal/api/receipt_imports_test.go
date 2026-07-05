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

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")

	// Test without captcha
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Error("expected no captcha_url when status is pending")
	}

	// Test with captcha
	ri.setStatus(StatusWaitingCaptcha)
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}
	expectedURL := "/receipts/imports/" + ri.ID + "/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %q", expectedURL, res["captcha_url"])
	}
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &mockFetcher{} })

	tests := []struct {
		name       string
		req        any
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
				"access_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access key",
			req: map[string]string{
				"access_key": "35240100000000000000650010000000010000000003", // 44 digits
			},
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			s.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
