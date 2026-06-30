package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{Page: []byte("<html></html>")}, nil
}

func TestCreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(captcha.Solver) sefaz.Fetcher {
		return &dummyFetcher{}
	})

	t.Run("InvalidAccessKey", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"access_key": "123",
		})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		err := s.createImport(rr, req)
		if err == nil {
			t.Fatal("expected error for invalid access key")
		}

		var he *apierrors.HTTPError
		if !errors.As(err, &he) || he.Status != http.StatusBadRequest {
			t.Errorf("expected BadRequest error, got %v", err)
		}
	})

	t.Run("ValidAccessKey", func(t *testing.T) {
		// 44 digits
		accessKey := "35240112345678000123650010000000011000000011"
		body, _ := json.Marshal(map[string]string{
			"access_key": accessKey,
		})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		err := s.createImport(rr, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rr.Code != http.StatusAccepted {
			t.Errorf("expected 202 Accepted, got %d", rr.Code)
		}
	})
}

func TestReceiptImport_MarshalJSON(t *testing.T) {
	ri := newReceiptImport("key", "url")

	t.Run("PendingNoCaptchaURL", func(t *testing.T) {
		ri.status = StatusPending
		data, err := json.Marshal(ri)
		if err != nil {
			t.Fatal(err)
		}
		var res map[string]any
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatal(err)
		}
		if _, ok := res["captcha_url"]; ok {
			t.Error("captcha_url should be omitted when not waiting_captcha")
		}
	})

	t.Run("WaitingCaptchaHasURL", func(t *testing.T) {
		ri.status = StatusWaitingCaptcha
		data, err := json.Marshal(ri)
		if err != nil {
			t.Fatal(err)
		}
		var res map[string]any
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatal(err)
		}
		url, ok := res["captcha_url"].(string)
		if !ok || url == "" {
			t.Error("captcha_url should be present when waiting_captcha")
		}
		expected := "/receipts/imports/" + ri.ID + "/captcha"
		if url != expected {
			t.Errorf("expected %q, got %q", expected, url)
		}
	})
}
