package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := &ReceiptImport{
		ID:     "123",
		status: StatusWaitingCaptcha,
	}

	data, err := ri.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}

	if res["captcha_url"] != "/receipts/imports/123/captcha" {
		t.Errorf("want captcha_url /receipts/imports/123/captcha, got %v", res["captcha_url"])
	}

	ri.status = StatusCompleted
	data, err = ri.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal(err)
	}
	if _, ok := res["captcha_url"]; ok {
		t.Error("captcha_url should be omitted when status is not waiting_captcha")
	}
}

type dummyFetcher struct{}

func (f *dummyFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return &sefaz.Result{}, nil
}

func TestServer_CreateImport_Validation(t *testing.T) {
	dummyFetcherFactory := func(captcha.Solver) sefaz.Fetcher { return &dummyFetcher{} }
	srv := NewServer(nil, dummyFetcherFactory)

	t.Run("InvalidAccessKey", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"access_key": "123"})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status 400, got %d", rr.Code)
		}

		var res map[string]any
		if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(res["error"].(string), "invalid access_key") {
			t.Errorf("expected error message to contain 'invalid access_key', got %v", res["error"])
		}
	})

	t.Run("ValidAccessKey", func(t *testing.T) {
		// 44 digits
		validKey := "12345678901234567890123456789012345678901234"
		body, _ := json.Marshal(map[string]string{"access_key": validKey})
		req := httptest.NewRequest("POST", "/receipts/imports", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusAccepted {
			t.Errorf("want status 202, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})
}
