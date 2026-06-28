package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, nil)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "exactly one of access_key or qr_url is required",
		},
		{
			name:       "invalid access key",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "access key must be 44 numeric digits",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			s.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tc.wantStatus)
			}

			var resp map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if msg, ok := resp["error"].(string); !ok || msg != tc.wantError {
				t.Errorf("got error %q, want %q", msg, tc.wantError)
			}
		})
	}
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
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

	wantURL := "/receipts/imports/test-id/captcha"
	if got := res["captcha_url"]; got != wantURL {
		t.Errorf("captcha_url = %q, want %q", got, wantURL)
	}

	// Verify it's omitted when not waiting_captcha
	ri.status = StatusCompleted
	b, _ = json.Marshal(ri)
	res = make(map[string]any)
	_ = json.Unmarshal(b, &res)
	if got, ok := res["captcha_url"]; ok {
		t.Errorf("expected captcha_url to be omitted, got %q", got)
	}
}
