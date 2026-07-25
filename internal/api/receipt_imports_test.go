package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type mockFetcher struct{}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func dummyFetcherFactory(solver captcha.Solver) sefaz.Fetcher {
	return &mockFetcher{}
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	// Test normal status without captcha_url
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")
	ri.status = StatusPending

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	res := make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Error("expected captcha_url to be omitted when status is pending")
	}

	// Test StatusWaitingCaptcha with captcha_url
	ri.status = StatusWaitingCaptcha
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	captchaURL, ok := res["captcha_url"].(string)
	if !ok || captchaURL != "/receipts/imports/"+ri.ID+"/captcha" {
		t.Errorf("expected captcha_url to be /receipts/imports/%s/captcha, got %v", ri.ID, res["captcha_url"])
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, dummyFetcherFactory)

	// Valid access key (44 digits)
	validBody, err := json.Marshal(map[string]string{
		"access_key": "29240100000000000000650010000000010000000003",
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(validBody))
	w := httptest.NewRecorder()

	err = s.createImport(w, req)
	if err != nil {
		t.Fatalf("expected no error for valid access key, got: %v", err)
	}

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	// Invalid access key format
	invalidBody, err := json.Marshal(map[string]string{
		"access_key": "short",
	})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	req = httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(invalidBody))
	w = httptest.NewRecorder()

	err = s.createImport(w, req)
	if err == nil {
		t.Fatal("expected error for invalid access key")
	}

	var he *apierrors.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}

	if he.Status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, he.Status)
	}

	expectedMsg := "access key must be 44 digits"
	if he.Msg != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, he.Msg)
	}
}

func TestReceiptImport_ReceiptIDOutput(t *testing.T) {
	receiptID := int64(12345)
	ri := newReceiptImport("29240100000000000000650010000000010000000003", "")
	ri.ReceiptID = &receiptID

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	res := make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	val, ok := res["receipt_id"]
	if !ok {
		t.Fatal("expected receipt_id to be present")
	}

	floatVal, ok := val.(float64)
	if !ok || int64(floatVal) != receiptID {
		t.Errorf("expected receipt_id to be %d, got %v", receiptID, val)
	}
}

type syncMockFetcher struct {
	done chan struct{}
}

func (f *syncMockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	close(f.done)
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestSubmitImport_Async(t *testing.T) {
	doneCh := make(chan struct{})
	fetcher := &syncMockFetcher{done: doneCh}

	s := NewServer(nil, func(solver captcha.Solver) sefaz.Fetcher {
		return fetcher
	})

	_ = s.submitImport("29240100000000000000650010000000010000000003", "")

	select {
	case <-doneCh:
		// Success: fetcher was invoked without busy waiting.
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for fetcher to be invoked")
	}
}
