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

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher { return nil })

	cases := []struct {
		name   string
		body   map[string]string
		status int
		errMsg string
	}{
		{
			name:   "missing both",
			body:   map[string]string{},
			status: http.StatusBadRequest,
			errMsg: "either access_key or qr_url is required",
		},
		{
			name: "both provided",
			body: map[string]string{
				"access_key": "12345678901234567890123456789012345678901234",
				"qr_url":     "http://example.com",
			},
			status: http.StatusBadRequest,
			errMsg: "access_key and qr_url are mutually exclusive",
		},
		{
			name:   "invalid access key length",
			body:   map[string]string{"access_key": "123"},
			status: http.StatusBadRequest,
			errMsg: "invalid access_key: must be 44 numeric digits",
		},
		{
			name:   "non-numeric access key",
			body:   map[string]string{"access_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
			status: http.StatusBadRequest,
			errMsg: "invalid access_key: must be 44 numeric digits",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)
			if err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tc.status {
				t.Errorf("status: want %d, got %d", tc.status, rr.Code)
			}

			var resp map[string]string
			_ = json.Unmarshal(rr.Body.Bytes(), &resp)
			if resp["error"] != tc.errMsg {
				t.Errorf("error: want %q, got %q", tc.errMsg, resp["error"])
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
		t.Fatalf("marshal: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := "/receipts/imports/test-id/captcha"
	if got := data["captcha_url"]; got != want {
		t.Errorf("captcha_url: want %q, got %q", want, got)
	}
}
