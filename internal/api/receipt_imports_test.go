package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReceiptImport_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		wantURL  bool
		wantPath string
	}{
		{
			name:    "pending - no captcha_url",
			status:  StatusPending,
			wantURL: false,
		},
		{
			name:     "waiting_captcha - has captcha_url",
			status:   StatusWaitingCaptcha,
			wantURL:  true,
			wantPath: "/captcha",
		},
		{
			name:    "completed - no captcha_url",
			status:  StatusCompleted,
			wantURL: false,
		},
		{
			name:    "failed - no captcha_url",
			status:  StatusFailed,
			wantURL: false,
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

			res := make(map[string]any)
			if err := json.Unmarshal(b, &res); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			url, ok := res["captcha_url"].(string)
			if tt.wantURL {
				if !ok {
					t.Errorf("MarshalJSON() missing captcha_url field")
				}
				if !strings.HasSuffix(url, tt.wantPath) {
					t.Errorf("captcha_url = %q, want suffix %q", url, tt.wantPath)
				}
				if !strings.Contains(url, ri.ID) {
					t.Errorf("captcha_url = %q, want it to contain ID %q", url, ri.ID)
				}
			} else {
				if ok && url != "" {
					t.Errorf("MarshalJSON() unexpected captcha_url field: %q", url)
				}
			}
		})
	}
}
