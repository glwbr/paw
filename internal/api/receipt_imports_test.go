package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		ri         *ReceiptImport
		wantStatus Status
		wantURL    string
	}{
		{
			name: "pending status",
			ri: &ReceiptImport{
				ID:     "123",
				status: StatusPending,
			},
			wantStatus: StatusPending,
			wantURL:    "",
		},
		{
			name: "waiting_captcha status",
			ri: &ReceiptImport{
				ID:     "123",
				status: StatusWaitingCaptcha,
			},
			wantStatus: StatusWaitingCaptcha,
			wantURL:    "/receipts/imports/123/captcha",
		},
		{
			name: "completed status",
			ri: &ReceiptImport{
				ID:     "123",
				status: StatusCompleted,
			},
			wantStatus: StatusCompleted,
			wantURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.ri.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got := res["status"]; got != string(tt.wantStatus) {
				t.Errorf("status = %v, want %v", got, tt.wantStatus)
			}

			gotURL, ok := res["captcha_url"].(string)
			if tt.wantURL == "" {
				if ok && gotURL != "" {
					t.Errorf("captcha_url = %v, want empty", gotURL)
				}
			} else {
				if gotURL != tt.wantURL {
					t.Errorf("captcha_url = %v, want %v", gotURL, tt.wantURL)
				}
			}
		})
	}
}
