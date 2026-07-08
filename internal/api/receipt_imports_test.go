package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

func TestReceiptImport_MarshalJSON_CaptchaURL(t *testing.T) {
	ri := newReceiptImport("123", "")
	ri.setStatus(StatusWaitingCaptcha)

	data, err := json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedURL := "/receipts/imports/" + ri.ID + "/captcha"
	if res["captcha_url"] != expectedURL {
		t.Errorf("expected captcha_url %q, got %q", expectedURL, res["captcha_url"])
	}

	ri.setStatus(StatusCompleted)
	data, err = json.Marshal(ri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	res = make(map[string]any)
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, ok := res["captcha_url"]; ok {
		t.Error("captcha_url should be omitted when status is not waiting_captcha")
	}
}

type mockFetcher struct {
	fn func(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error)
}

func (f *mockFetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	return f.fn(ctx, req)
}

func TestServer_CreateImport_Validation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{
			fn: func(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
				return &sefaz.Result{Page: []byte("<html></html>")}, nil
			},
		}
	})

	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "invalid access key length",
			body:       `{"access_key": "123"}`,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "access key must be 44 numeric digits",
		},
		{
			name:       "invalid access key characters",
			body:       `{"access_key": "4444444444444444444444444444444444444444444a"}`,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "access key must be 44 numeric digits",
		},
		{
			name:       "valid access key",
			body:       `{"access_key": "29240112345678000199650010000000011000000010"}`,
			wantStatus: http.StatusAccepted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			s.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantMsg != "" {
				var res map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if !strings.Contains(res["error"], tc.wantMsg) {
					t.Errorf("error message: want %q to contain %q", res["error"], tc.wantMsg)
				}
			}
		})
	}
}

func TestWriteError_Mapping(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Test ErrNotFound mapping
	writeError(rec, req, ErrNotFound)
	if rec.Code != http.StatusNotFound {
		t.Errorf("ErrNotFound: want 404, got %d", rec.Code)
	}

	// Test ErrCaptchaNotPending mapping
	rec = httptest.NewRecorder()
	writeError(rec, req, ErrCaptchaNotPending)
	if rec.Code != http.StatusConflict {
		t.Errorf("ErrCaptchaNotPending: want 409, got %d", rec.Code)
	}

	// Test apierrors.HTTPError mapping
	rec = httptest.NewRecorder()
	writeError(rec, req, apierrors.BadRequest("custom message", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("HTTPError: want 400, got %d", rec.Code)
	}
	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res["error"] != "custom message" {
		t.Errorf("HTTPError message: want %q, got %q", "custom message", res["error"])
	}
}

func TestServer_CreateImport_LongRunningOperation(t *testing.T) {
	s := NewServer(nil, func(s captcha.Solver) sefaz.Fetcher {
		return &mockFetcher{
			fn: func(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
				return &sefaz.Result{Page: []byte("<html></html>")}, nil
			},
		}
	})

	body := `{"access_key": "29240112345678000199650010000000011000000010"}`
	req := httptest.NewRequest("POST", "/receipts/imports", strings.NewReader(body))
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status: want 202, got %d", rec.Code)
	}

	loc := rec.Header().Get("Operation-Location")
	if !strings.HasPrefix(loc, "/receipts/imports/") {
		t.Errorf("Operation-Location: want prefix /receipts/imports/, got %q", loc)
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if !strings.HasSuffix(loc, res.ID) {
		t.Errorf("Operation-Location %q should end with ID %q", loc, res.ID)
	}
}

func TestReceiptImport_Solve_Blocks(t *testing.T) {
	ri := newReceiptImport("123", "")
	ch := captcha.Challenge{ID: "ch1", Image: []byte("img"), ContentType: "image/png"}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		solution, err := ri.Solve(context.Background(), ch)
		if err != nil {
			t.Errorf("Solve failed: %v", err)
		}
		if solution.Text != "ans123" {
			t.Errorf("Solve: want text ans123, got %q", solution.Text)
		}
	}()

	// Wait for status change with a timeout
	deadline := time.Now().Add(time.Second)
	for ri.Status() != StatusWaitingCaptcha {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for status waiting_captcha")
		}
		time.Sleep(time.Millisecond)
	}

	if err := ri.submitCaptcha("ans123"); err != nil {
		t.Errorf("submitCaptcha failed: %v", err)
	}

	wg.Wait()
}
