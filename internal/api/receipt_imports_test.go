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
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got, ok := res["captcha_url"].(string)
	if !ok {
		t.Errorf("missing captcha_url in JSON")
	}
	want := "/receipts/imports/test-id/captcha"
	if got != want {
		t.Errorf("captcha_url = %q, want %q", got, want)
	}

	// Test StatusCompleted (should not have captcha_url)
	ri.status = StatusCompleted
	b, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be missing for status %s", ri.status)
	}
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{}
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "invalid access key",
			body:       map[string]string{"access_key": "invalid"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid access key",
			body:       map[string]string{"access_key": "29240112345678000100650010000000011000000009"}, // 44 digits
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			if err := s.createImport(rr, req); err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
