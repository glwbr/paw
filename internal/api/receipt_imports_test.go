package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.status = StatusWaitingCaptcha

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}

	got, ok := res["captcha_url"].(string)
	if !ok {
		t.Error("missing captcha_url in JSON")
	}
	want := "/receipts/imports/" + ri.ID + "/captcha"
	if got != want {
		t.Errorf("captcha_url = %q, want %q", got, want)
	}
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return nil })

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid access key length",
			body:       map[string]string{"access_key": "123"},
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid access key format",
		},
		{
			name:       "invalid access key characters",
			body:       map[string]string{"access_key": "1234567890123456789012345678901234567890123a"},
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid access key format",
		},
		{
			name:       "valid access key",
			body:       map[string]string{"access_key": "12345678901234567890123456789012345678901234"},
			wantStatus: http.StatusAccepted,
		},
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
				json.Unmarshal(rr.Body.Bytes(), &res)
				if got := res["error"].(string); got != tt.wantErr {
					t.Errorf("error = %q, want %q", got, tt.wantErr)
				}
			}
		})
	}
}
