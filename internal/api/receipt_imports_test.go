package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.ID = "test-id"

	// Case 1: StatusPending
	ri.status = StatusPending
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("expected captcha_url to be omitted in status pending, got %v", res["captcha_url"])
	}

	// Case 2: StatusWaitingCaptcha
	res = make(map[string]any) // always re-initialize the target map before unmarshaling to avoid stale data
	ri.status = StatusWaitingCaptcha
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	gotURL, ok := res["captcha_url"]
	if !ok {
		t.Errorf("expected captcha_url to be present in status waiting_captcha")
	} else if gotURL != "/receipts/imports/test-id/captcha" {
		t.Errorf("expected captcha_url /receipts/imports/test-id/captcha, got %v", gotURL)
	}
}
