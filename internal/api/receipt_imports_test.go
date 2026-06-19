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
			ri := newReceiptImport("123", "")
			ri.status = tt.status

			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			_, ok := res["captcha_url"]
			if ok != tt.wantCaptcha {
				t.Errorf("got captcha_url=%v, want %v", ok, tt.wantCaptcha)
			}

			if tt.wantCaptcha {
				want := "/receipts/imports/" + ri.ID + "/captcha"
				if res["captcha_url"] != want {
					t.Errorf("got captcha_url=%v, want %v", res["captcha_url"], want)
				}
			}
		})
	}
}
