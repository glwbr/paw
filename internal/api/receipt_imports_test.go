package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	data, err := ri.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got, ok := res["captcha_url"].(string)
	if !ok {
		t.Errorf("expected captcha_url to be present, but it was missing or not a string")
	}
	want := "/receipts/imports/test-id/captcha"
	if got != want {
		t.Errorf("captcha_url: want %q, got %q", want, got)
	}
}

func TestCreateImport_Validation(t *testing.T) {
	dummyFetcherFactory := func(captcha.Solver) sefaz.Fetcher {
		return nil
	}
	s := NewServer(nil, dummyFetcherFactory)

	// Invalid access key (too short)
	reqBody := `{"access_key": "123"}`
	req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	err := s.createImport(w, req)

	if err == nil {
		t.Errorf("expected validation error for invalid access key, got nil")
	}
}
