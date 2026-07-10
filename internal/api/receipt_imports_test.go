package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		status  Status
		wantURL bool
	}{
		{
			name:    "pending has no captcha_url",
			status:  StatusPending,
			wantURL: false,
		},
		{
			name:    "waiting_captcha has captcha_url",
			status:  StatusWaitingCaptcha,
			wantURL: true,
		},
		{
			name:    "completed has no captcha_url",
			status:  StatusCompleted,
			wantURL: false,
		},
		{
			name:    "failed has no captcha_url",
			status:  StatusFailed,
			wantURL: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ri := &ReceiptImport{
				ID:     "test-id",
				status: tc.status,
			}

			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tc.wantURL {
				if !ok {
					t.Error("expected captcha_url field, got none")
				}
				expected := "/receipts/imports/" + ri.ID + "/captcha"
				if url != expected {
					t.Errorf("expected captcha_url %q, got %q", expected, url)
				}
			} else {
				if ok {
					t.Errorf("unexpected captcha_url field: %q", url)
				}
			}

			// Verify other required fields are present
			for _, field := range []string{"id", "status", "done", "created_at", "updated_at"} {
				if _, ok := res[field]; !ok {
					t.Errorf("missing required field in JSON: %q", field)
				}
			}

			// Verify 'done' boolean logic
			done, _ := res["done"].(bool)
			if done != tc.status.Terminal() {
				t.Errorf("expected done=%v for status %q, got %v", tc.status.Terminal(), tc.status, done)
			}
		})
	}
}

func TestStatusTerminal(t *testing.T) {
	cases := []struct {
		status Status
		want   bool
	}{
		{StatusPending, false},
		{StatusFetching, false},
		{StatusWaitingCaptcha, false},
		{StatusProcessing, false},
		{StatusCompleted, true},
		{StatusFailed, true},
	}
	for _, tc := range cases {
		if got := tc.status.Terminal(); got != tc.want {
			t.Errorf("%s.Terminal() = %v, want %v", tc.status, got, tc.want)
		}
	}
}
