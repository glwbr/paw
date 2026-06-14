package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON_Hypermedia(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		wantURL string
	}{
		{
			name:   "pending",
			status: StatusPending,
		},
		{
			name:    "waiting_captcha",
			status:  StatusWaitingCaptcha,
			wantURL: "/receipts/imports/123/captcha",
		},
		{
			name:   "completed",
			status: StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri := &ReceiptImport{ID: "123", status: tt.status}
			b, _ := json.Marshal(ri)
			var res map[string]any
			_ = json.Unmarshal(b, &res)

			url, _ := res["captcha_url"].(string)
			if url != tt.wantURL {
				t.Errorf("got captcha_url %q, want %q", url, tt.wantURL)
			}
		})
	}
}
