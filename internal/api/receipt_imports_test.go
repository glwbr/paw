package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	cases := []struct {
		name       string
		status     Status
		id         string
		wantURL    bool
		wantURLVal string
	}{
		{
			name:       "waiting_captcha includes captcha_url",
			status:     StatusWaitingCaptcha,
			id:         "123",
			wantURL:    true,
			wantURLVal: "/receipts/imports/123/captcha",
		},
		{
			name:    "pending omits captcha_url",
			status:  StatusPending,
			id:      "123",
			wantURL: false,
		},
		{
			name:    "completed omits captcha_url",
			status:  StatusCompleted,
			id:      "123",
			wantURL: false,
		},
		{
			name:    "failed omits captcha_url",
			status:  StatusFailed,
			id:      "123",
			wantURL: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ri := &ReceiptImport{
				ID:     tc.id,
				status: tc.status,
			}

			b, err := json.Marshal(ri)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			val, ok := res["captcha_url"]
			if tc.wantURL {
				if !ok {
					t.Error("missing captcha_url in JSON")
				}
				if val != tc.wantURLVal {
					t.Errorf("captcha_url: want %q, got %q", tc.wantURLVal, val)
				}
			} else {
				if ok {
					t.Errorf("unexpected captcha_url in JSON: %q", val)
				}
			}

			// Also verify "done" field
			done, ok := res["done"].(bool)
			if !ok {
				t.Error("missing or invalid done field in JSON")
			}
			if done != tc.status.Terminal() {
				t.Errorf("done: want %v, got %v", tc.status.Terminal(), done)
			}
		})
	}
}

func TestStatus_Terminal(t *testing.T) {
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
			t.Errorf("%s.Terminal(): want %v, got %v", tc.status, tc.want, got)
		}
	}
}
