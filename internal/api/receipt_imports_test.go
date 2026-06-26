package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "test-id",
		status: StatusWaitingCaptcha,
	}

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedURL := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %q", expectedURL, res["captcha_url"])
	}

	// Test when not waiting captcha
	ri.status = StatusPending
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted when not waiting captcha")
	}
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := &Server{} // Minimal server for validation check

	// Invalid access key (too short)
	req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(`{"access_key": "123"}`))
	rr := httptest.NewRecorder()

	err := s.createImport(rr, req)
	if err == nil {
		t.Fatal("expected error for invalid access key")
	}

	var he *apierrors.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if he.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", he.Status)
	}
	if he.Msg != "invalid access key" {
		t.Errorf("expected msg 'invalid access key', got %q", he.Msg)
	}
}
