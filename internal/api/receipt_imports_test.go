package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     Status
		wantCaptcha bool
	}{
		{
			name:       "pending",
			status:     StatusPending,
			wantCaptcha: false,
		},
		{
			name:       "waiting_captcha",
			status:     StatusWaitingCaptcha,
			wantCaptcha: true,
		},
		{
			name:       "completed",
			status:     StatusCompleted,
			wantCaptcha: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ri := &ReceiptImport{
				ID:     "test-id",
				status: tt.status,
			}

			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tt.wantCaptcha {
				if !ok {
					t.Errorf("expected captcha_url to be present")
				}
				want := "/receipts/imports/test-id/captcha"
				if url != want {
					t.Errorf("captcha_url = %q, want %q", url, want)
				}
			} else {
				if ok && url != "" {
					t.Errorf("expected captcha_url to be absent, got %q", url)
				}
			}
		})
	}
}
