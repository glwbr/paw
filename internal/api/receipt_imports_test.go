package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
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
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if res["id"] != "test-id" {
		t.Errorf("expected id 'test-id', got %v", res["id"])
	}

	if res["status"] != string(StatusWaitingCaptcha) {
		t.Errorf("expected status %s, got %v", StatusWaitingCaptcha, res["status"])
	}

	expectedURL := "/receipts/imports/test-id/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %v", expectedURL, res["captcha_url"])
	}

	// Test status without captcha_url
	ri.status = StatusFetching
	data, err = ri.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should be omitted for status %s", StatusFetching)
	}
}
