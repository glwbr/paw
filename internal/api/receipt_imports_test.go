package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		status   Status
		wantURL  string
		expected bool
	}{
		{
			name:     "pending",
			status:   StatusPending,
			expected: false,
		},
		{
			name:     "waiting_captcha",
			status:   StatusWaitingCaptcha,
			wantURL:  "/receipts/imports/123/captcha",
			expected: true,
		},
		{
			name:     "fetching",
			status:   StatusFetching,
			expected: false,
		},
		{
			name:     "processing",
			status:   StatusProcessing,
			expected: false,
		},
		{
			name:     "completed",
			status:   StatusCompleted,
			expected: false,
		},
		{
			name:     "failed",
			status:   StatusFailed,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ReceiptImport{
				ID:     "123",
				status: tc.status,
			}

			data, err := json.Marshal(r)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(data, &res); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}

			val, ok := res["captcha_url"]
			if tc.expected {
				if !ok {
					t.Error("expected captcha_url to be present")
				}
				if val != tc.wantURL {
					t.Errorf("captcha_url = %q, want %q", val, tc.wantURL)
				}
			} else {
				if ok {
					t.Errorf("expected captcha_url to be omitted, got %q", val)
				}
			}
		})
	}
}
