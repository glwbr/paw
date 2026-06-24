package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.status = StatusWaitingCaptcha

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if res["captcha_url"] != "/receipts/imports/"+ri.ID+"/captcha" {
		t.Errorf("expected captcha_url, got %v", res["captcha_url"])
	}

	ri.status = StatusCompleted
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Errorf("captcha_url should not be present when not waiting for captcha")
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(solver captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	t.Run("invalid access key format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"access_key": "short",
		})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		err := s.createImport(rr, req)
		if err == nil {
			t.Errorf("expected error for invalid access key")
		}
	})

	t.Run("valid access key format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"access_key": "12345678901234567890123456789012345678901234",
		})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		err := s.createImport(rr, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if rr.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", rr.Code)
		}
	})
}
