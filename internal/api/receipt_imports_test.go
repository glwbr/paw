package api

import (
	"bytes"
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
	return &sefaz.Result{}, nil
}

func TestCreateImport_InvalidAccessKey(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	body := `{"access_key": "invalid"}`
	req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader([]byte(body)))
	rr := httptest.NewRecorder()

	err := s.createImport(rr, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var he *apierrors.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if he.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", he.Status)
	}

	if !strings.Contains(he.Msg, "invalid access key") {
		t.Errorf("expected error message to contain 'invalid access key', got %q", he.Msg)
	}
}

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("29240100000000000000650010000000010000000000", "")
	ri.status = StatusWaitingCaptcha

	data, err := ri.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal ReceiptImport: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	got, ok := res["captcha_url"].(string)
	if !ok {
		t.Fatal("expected captcha_url field in JSON")
	}

	want := "/receipts/imports/" + ri.ID + "/captcha"
	if got != want {
		t.Errorf("expected captcha_url %q, got %q", want, got)
	}

	// Verify it's omitted when not waiting for captcha
	ri.status = StatusPending
	data, _ = ri.MarshalJSON()
	res = make(map[string]any)
	_ = json.Unmarshal(data, &res)
	if _, ok := res["captcha_url"]; ok {
		t.Error("expected captcha_url to be omitted when status is pending")
	}
}
