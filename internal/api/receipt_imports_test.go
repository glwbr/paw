package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	rid := int64(123)
	tests := []struct {
		name string
		ri   *ReceiptImport
		want string
	}{
		{
			name: "waiting_captcha",
			ri:   &ReceiptImport{ID: "abc", status: StatusWaitingCaptcha},
			want: `"captcha_url":"/receipts/imports/abc/captcha"`,
		},
		{
			name: "completed",
			ri:   &ReceiptImport{ID: "abc", status: StatusCompleted, ReceiptID: &rid},
			want: `"receipt_id":123`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.ri)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(data, []byte(tt.want)) {
				t.Errorf("MarshalJSON() = %s, want to contain %s", data, tt.want)
			}
		})
	}
}

type mockFetcher struct{}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return &mockFetcher{} })
	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantErr    string
	}{
		{"invalid length", map[string]string{"access_key": "123"}, 400, "invalid access key format"},
		{"invalid chars", map[string]string{"access_key": "1234567890123456789012345678901234567890123a"}, 400, "invalid access key format"},
		{"valid key", map[string]string{"access_key": "12345678901234567890123456789012345678901234"}, 202, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()
			s.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantErr != "" {
				var res map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if got := res["error"].(string); got != tt.wantErr {
					t.Errorf("error = %q, want %q", got, tt.wantErr)
				}
			}
		})
	}
}
