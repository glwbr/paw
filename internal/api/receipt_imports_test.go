package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReceiptImport_CaptchaURL(t *testing.T) {
	s := NewServer(nil, nil)
	ri := newReceiptImport("29240112345678901234650010000000012345678901", "")
	s.mu.Lock()
	s.imports[ri.ID] = ri
	s.mu.Unlock()

	// 1. Initially (pending), captcha_url should be empty
	req := httptest.NewRequest("GET", "/receipts/imports/"+ri.ID, nil)
	req.SetPathValue("id", ri.ID)
	rr := httptest.NewRecorder()
	if err := s.getImport(rr, req); err != nil {
		t.Fatalf("getImport failed: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("expected no captcha_url in pending status, got %v", res["captcha_url"])
	}

	// 2. Flip to waiting_captcha, captcha_url should be present
	ri.mu.Lock()
	ri.status = StatusWaitingCaptcha
	ri.mu.Unlock()

	rr = httptest.NewRecorder()
	if err := s.getImport(rr, req); err != nil {
		t.Fatalf("getImport failed: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	expectedURL := "/receipts/imports/" + ri.ID + "/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %q", expectedURL, res["captcha_url"])
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, nil)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "invalid access key format",
			body:       map[string]string{"access_key": "short"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid access key format",
			body:       map[string]string{"access_key": "29240112345678901234650010000000012345678901"},
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			err := s.createImport(rr, req)
			if err != nil {
				writeError(rr, req, err)
			}

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
