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
			name:       "invalid access key format",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid access_key format: must be 44 digits",
		},
		{
			name:       "invalid access key characters",
			body:       `{"access_key": "1234567890123456789012345678901234567890123a"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid access_key format: must be 44 digits",
		},
		{
			name:       "missing both fields",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "exactly one of access_key or qr_url is required",
		},
		{
			name:       "both fields provided",
			body:       `{"access_key": "12345678901234567890123456789012345678901234", "qr_url": "http://example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "exactly one of access_key or qr_url is required",
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

			var resp map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}

			if resp["error"] != tc.wantError {
				t.Errorf("got error %q, want %q", resp["error"], tc.wantError)
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
		t.Errorf("captcha_url = %v, want %v", got, want)
	}
}

func TestReceiptImport_MarshalJSON_NoCaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusPending,
	}

	b, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]any
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted when status is not waiting_captcha")
	}
}
