package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		wantURL bool
	}{
		{
			name:    "Pending - no URL",
			status:  StatusPending,
			wantURL: false,
		},
		{
			name:    "WaitingCaptcha - has URL",
			status:  StatusWaitingCaptcha,
			wantURL: true,
		},
		{
			name:    "Completed - no URL",
			status:  StatusCompleted,
			wantURL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri := &ReceiptImport{
				ID:     "test-id",
				status: tt.status,
			}

			data, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(data, &res); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			url, ok := res["captcha_url"]
			if tt.wantURL {
				if !ok {
					t.Error("expected captcha_url field, but it was missing")
				} else if url != "/receipts/imports/test-id/captcha" {
					t.Errorf("expected captcha_url %q, got %q", "/receipts/imports/test-id/captcha", url)
				}
			} else {
				if ok {
					t.Errorf("did not expect captcha_url field, but got %v", url)
				}
			}
		})
	}
}
