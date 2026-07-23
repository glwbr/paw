package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func dummyFetcherFactory(captcha.Solver) sefaz.Fetcher {
	return &dummyFetcher{}
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("12345678901234567890123456789012345678901234", "")

	// Test case 1: StatusPending (should NOT have captcha_url)
	ri.status = StatusPending
	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	res := make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, exists := res["captcha_url"]; exists {
		t.Error("expected captcha_url to be omitted in StatusPending")
	}

	// Test case 2: StatusWaitingCaptcha (should HAVE captcha_url)
	ri.status = StatusWaitingCaptcha
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	val, exists := res["captcha_url"]
	if !exists {
		t.Error("expected captcha_url to be present in StatusWaitingCaptcha")
	} else if val != "/receipts/imports/"+ri.ID+"/captcha" {
		t.Errorf("expected captcha_url to be /receipts/imports/%s/captcha, got %v", ri.ID, val)
	}
}

func TestReceiptImport_ReceiptIDSerialization(t *testing.T) {
	id := int64(42)
	ri := &ReceiptImport{
		ID:        "test-import-id",
		status:    StatusCompleted,
		ReceiptID: &id,
	}

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got, ok := res["receipt_id"]
	if !ok {
		t.Fatal("expected receipt_id in JSON, but it was missing")
	}

	gotFloat, ok := got.(float64)
	if !ok {
		t.Fatalf("expected receipt_id to be a number, got %T", got)
	}

	if int64(gotFloat) != 42 {
		t.Errorf("expected receipt_id to be 42, got %v", gotFloat)
	}
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, dummyFetcherFactory)

	// Test case 1: Invalid access key (length mismatch)
	reqBody := `{"access_key": "123"}`
	req := httptest.NewRequest(http.MethodPost, "/receipts/imports", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	err := s.createImport(w, req)
	if err == nil {
		t.Fatal("expected validation error for invalid access key length, got nil")
	}

	var he *apierrors.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}

	if he.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", he.Status)
	}

	if he.Msg != "access key must be 44 numeric digits" {
		t.Errorf("expected public message 'access key must be 44 numeric digits', got %q", he.Msg)
	}

	// Test case 2: Valid access key (should succeed with 202)
	validKey := "29240100000000000000650010000000010000000009"
	reqBody = `{"access_key": "` + validKey + `"}`
	req = httptest.NewRequest(http.MethodPost, "/receipts/imports", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	err = s.createImport(w, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", resp.StatusCode)
	}

	opLoc := resp.Header.Get("Operation-Location")
	if !strings.HasPrefix(opLoc, "/receipts/imports/") {
		t.Errorf("expected Operation-Location header, got %q", opLoc)
	}
}
