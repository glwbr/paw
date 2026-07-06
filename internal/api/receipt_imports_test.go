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

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal ReceiptImport: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal ReceiptImport: %v", err)
	}

	if res["captcha_url"] != "/receipts/imports/test-id/captcha" {
		t.Errorf("expected captcha_url to be /receipts/imports/test-id/captcha, got %v", res["captcha_url"])
	}

	ri.status = StatusCompleted
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal ReceiptImport: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal ReceiptImport: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Error("expected captcha_url to be omitted when status is not waiting_captcha")
	}
}
