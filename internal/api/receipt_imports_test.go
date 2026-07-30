package api

import (
	"encoding/json"
	"testing"
	"time"
)

// TestReceiptImportMarshalJSON tests that MarshalJSON correctly includes the conditional
// captcha_url hypermedia link when the import is in StatusWaitingCaptcha state, and
// omits it otherwise.
func TestReceiptImportMarshalJSON(t *testing.T) {
	ri := &ReceiptImport{
		ID:        "test-id",
		AccessKey: "12345678901234567890123456789012345678901234",
		status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test case 1: StatusPending (should omit captcha_url)
	var res map[string]any
	res = make(map[string]any)
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("expected captcha_url to be omitted in StatusPending")
	}

	// Test case 2: StatusWaitingCaptcha (should include captcha_url)
	ri.status = StatusWaitingCaptcha
	res = make(map[string]any)
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	captchaURL, ok := res["captcha_url"].(string)
	if !ok {
		t.Errorf("expected captcha_url to be present in StatusWaitingCaptcha")
	}
	expected := "/receipts/imports/test-id/captcha"
	if captchaURL != expected {
		t.Errorf("expected captcha_url to be %q, got %q", expected, captchaURL)
	}
}
