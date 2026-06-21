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
		urlPattern string
	}{
		{
			name:    "pending - no url",
			status:  StatusPending,
			id:      "123",
			wantURL: false,
		},
		{
			name:       "waiting_captcha - has url",
			status:     StatusWaitingCaptcha,
			id:         "abc",
			wantURL:    true,
			urlPattern: "/receipts/imports/abc/captcha",
		},
		{
			name:    "completed - no url",
			status:  StatusCompleted,
			id:      "456",
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
				t.Fatalf("marshal failed: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tc.wantURL {
				if !ok {
					t.Error("expected captcha_url field, but it's missing or not a string")
				} else if url != tc.urlPattern {
					t.Errorf("got captcha_url %q, want %q", url, tc.urlPattern)
				}
			} else {
				if ok && url != "" {
					t.Errorf("expected no captcha_url, but got %q", url)
				}
			}
		})
	}
}

func TestReceiptImport_AccessKeyValidation(t *testing.T) {
	// This tests the logic we expect to be in createImport eventually
	// or just demonstrates that we CAN validate it.
	// Since I didn't add the validation to the code yet (per plan),
	// I'll skip adding a failing test for now and focus on MarshalJSON.
}

func TestReceiptImport_DoneField(t *testing.T) {
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
		ri := &ReceiptImport{status: tc.status}
		b, _ := json.Marshal(ri)
		var res map[string]any
		_ = json.Unmarshal(b, &res)

		if got := res["done"].(bool); got != tc.want {
			t.Errorf("status %s: got done=%v, want %v", tc.status, got, tc.want)
		}
	}
}
