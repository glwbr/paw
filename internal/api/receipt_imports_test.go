package api

import (
	"encoding/json"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	cases := []struct {
		name       string
		status     Status
		wantURL    bool
		urlPattern string
	}{
		{
			name:    "pending - no captcha url",
			status:  StatusPending,
			wantURL: false,
		},
		{
			name:       "waiting_captcha - includes captcha url",
			status:     StatusWaitingCaptcha,
			wantURL:    true,
			urlPattern: "/receipts/imports/[a-f0-9]+/captcha",
		},
		{
			name:    "completed - no captcha url",
			status:  StatusCompleted,
			wantURL: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ri := newReceiptImport("123", "")
			ri.status = tc.status

			data, err := ri.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var res map[string]any
			if err := json.Unmarshal(data, &res); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tc.wantURL {
				if !ok {
					t.Error("expected captcha_url field, but it was missing or not a string")
				}
				if url == "" {
					t.Error("expected non-empty captcha_url")
				}
			} else {
				if ok && url != "" {
					t.Errorf("expected no captcha_url, but got %q", url)
				}
			}
		})
	}
}
