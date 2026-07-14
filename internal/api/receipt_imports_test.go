package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON_Hypermedia(t *testing.T) {
	ri := &ReceiptImport{ID: "test-id", status: StatusPending}

	// Case 1: Pending (no captcha_url)
	b, _ := ri.MarshalJSON()
	var res map[string]any
	_ = json.Unmarshal(b, &res)
	if _, ok := res["captcha_url"]; ok {
		t.Error("expected no captcha_url for status pending")
	}

	// Case 2: WaitingCaptcha (has captcha_url)
	ri.status = StatusWaitingCaptcha
	b, _ = ri.MarshalJSON()
	res = make(map[string]any)
	_ = json.Unmarshal(b, &res)
	want := "/receipts/imports/test-id/captcha"
	if got := res["captcha_url"]; got != want {
		t.Errorf("captcha_url = %q, want %q", got, want)
	}
}
