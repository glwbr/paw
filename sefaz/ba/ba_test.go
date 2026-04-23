package ba

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

// mockSolver returns a fixed solution or error.
type mockSolver struct {
	solutions []string // one per call; cycles the last entry
	err       error
	calls     int
}

func (m *mockSolver) Solve(_ context.Context, _ captcha.Challenge) (captcha.Solution, error) {
	if m.err != nil {
		return captcha.Solution{}, m.err
	}
	idx := m.calls
	if idx >= len(m.solutions) {
		idx = len(m.solutions) - 1
	}
	m.calls++
	return captcha.Solution{Text: m.solutions[idx], ChallengeID: "test"}, nil
}

const formStateHTML = `<input type="hidden" name="__VIEWSTATE" value="dmlld3N0YXRl">
<input type="hidden" name="__VIEWSTATEGENERATOR" value="ABCD1234">
<input type="hidden" name="__EVENTVALIDATION" value="ZXZlbnR2YWw=">`

// newTestPortal creates an httptest server simulating the BA portal.
// It handles the full flow: access key page, captcha, submission, tabs navigation, and print page.
func newTestPortal(t *testing.T, validCaptcha string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Captcha image
		case strings.Contains(r.URL.Path, "AntiRobo"):
			// Minimal valid JPEG so http.DetectContentType returns image/jpeg.
			_, _ = w.Write([]byte("\xff\xd8\xff\xe0JFIF"))

		// Access key page (GET = load form, POST = submit)
		case strings.Contains(r.URL.Path, "consulta_chave_acesso"):
			if r.Method == http.MethodGet {
				fmt.Fprintf(w, `<html><body>%s</body></html>`, formStateHTML)
				return
			}
			// POST: check captcha
			_ = r.ParseForm()
			if r.FormValue(fieldCaptcha) != validCaptcha {
				_, _ = fmt.Fprint(w, `<html><body>captcha inválido</body></html>`)
				return
			}
			// Success: return DANFE
			fmt.Fprintf(w, `<html><body>DANFE content%s</body></html>`, formStateHTML)

		// DANFE page (POST = "Visualizar em Abas")
		case strings.Contains(r.URL.Path, "consulta_danfe"):
			fmt.Fprintf(w, `<html><body>nfe tab content%s</body></html>`, formStateHTML)

		// Print page (GET)
		case strings.Contains(r.URL.Path, "Frm_Imprimir_parcial"):
			_, _ = fmt.Fprint(w, `<html><body>print page content</body></html>`)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
}

func newFetcherForTest(solver captcha.Solver) *Fetcher {
	return New(WithSolver(solver))
}

func TestFetcher_Fetch_qrURL(t *testing.T) {
	t.Parallel()

	srv := newTestPortal(t, "")
	defer srv.Close()

	// Override the portal to return DANFE on a GET to the QR URL
	qrSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.RawQuery, "qr=") {
			fmt.Fprintf(w, `<html><body>DANFE from QR%s</body></html>`, formStateHTML)
			return
		}
		// Forward all other requests to the main portal
		srv.Config.Handler.ServeHTTP(w, r)
	}))
	defer qrSrv.Close()

	// baseURL is a const; test the session directly against the test server.
	f := newFetcherForTest(nil)
	client, _ := f.newClient()
	s := &session{client: client, limiter: f.limiter}

	danfe, err := s.fetchQR(context.Background(), qrSrv.URL+"/qr?qr=1")
	if err != nil {
		t.Fatalf("fetchQR() error: %v", err)
	}
	if !strings.Contains(string(danfe), "DANFE from QR") {
		t.Errorf("expected DANFE content, got: %s", truncate(string(danfe), 100))
	}
}

func TestFetcher_Fetch_missingInput(t *testing.T) {
	t.Parallel()

	f := New()
	_, err := f.Fetch(context.Background(), &sefaz.Request{})
	if !errors.Is(err, sefaz.ErrMissingInput) {
		t.Errorf("expected ErrMissingInput, got: %v", err)
	}
}

func TestFetcher_Fetch_ambiguousInput(t *testing.T) {
	t.Parallel()

	f := New()
	_, err := f.Fetch(context.Background(), &sefaz.Request{
		AccessKey: "12345678901234567890123456789012345678901234",
		QRURL:     "https://example.com/qr",
	})
	if !errors.Is(err, sefaz.ErrAmbiguousInput) {
		t.Errorf("expected ErrAmbiguousInput, got: %v", err)
	}
}

func TestFetcher_Fetch_noCaptchaSolver(t *testing.T) {
	t.Parallel()

	f := New() // no solver
	_, err := f.Fetch(context.Background(), &sefaz.Request{
		AccessKey: "12345678901234567890123456789012345678901234",
	})
	if !errors.Is(err, sefaz.ErrNoCaptchaSolver) {
		t.Errorf("expected ErrNoCaptchaSolver, got: %v", err)
	}
}

func TestCheckErrors_patterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		html string
		want error
	}{
		{
			name: "captcha invalid",
			html: `<html><body>Captcha Inválido</body></html>`,
			want: sefaz.ErrCaptchaInvalid,
		},
		{
			name: "captcha retry message",
			html: `<html><body>Código incorreto tente novamente</body></html>`,
			want: sefaz.ErrCaptchaInvalid,
		},
		{
			name: "invalid access key",
			html: `<html><body>Chave de acesso inválida</body></html>`,
			want: &sefaz.InvalidAccessKeyError{},
		},
		{
			name: "invoice not found",
			html: `<html><body>NFC-e não encontrada</body></html>`,
			want: &sefaz.InvoiceNotFoundError{},
		},
		{
			name: "session expired",
			html: `<html><body>Sessão expirada</body></html>`,
			want: &sefaz.SessionExpiredError{},
		},
		{
			name: "generic error",
			html: `<html><body>Ocorreu um erro no servidor</body></html>`,
			want: &sefaz.PortalError{},
		},
		{
			name: "asp.net error",
			html: `<html><body>Object reference not set to an instance</body></html>`,
			want: &sefaz.PortalError{},
		},
		{
			name: "no error",
			html: `<html><body>Normal content</body></html>`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := checkErrors([]byte(tt.html))

			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil error, got: %v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("expected error %v, got nil", tt.want)
			}

			if sentinel, ok := tt.want.(interface{ Is(error) bool }); ok {
				_ = sentinel
				if !errors.Is(got, tt.want) {
					t.Errorf("expected errors.Is match for %v, got: %v", tt.want, got)
				}
				return
			}

			switch tt.want.(type) {
			case *sefaz.InvalidAccessKeyError:
				var target *sefaz.InvalidAccessKeyError
				if !errors.As(got, &target) {
					t.Errorf("expected InvalidAccessKeyError, got: %T %v", got, got)
				}
			case *sefaz.InvoiceNotFoundError:
				var target *sefaz.InvoiceNotFoundError
				if !errors.As(got, &target) {
					t.Errorf("expected InvoiceNotFoundError, got: %T %v", got, got)
				}
			case *sefaz.SessionExpiredError:
				var target *sefaz.SessionExpiredError
				if !errors.As(got, &target) {
					t.Errorf("expected SessionExpiredError, got: %T %v", got, got)
				}
			case *sefaz.PortalError:
				var target *sefaz.PortalError
				if !errors.As(got, &target) {
					t.Errorf("expected PortalError, got: %T %v", got, got)
				}
			}
		})
	}
}

func TestSession_accessKeyFlow(t *testing.T) {
	t.Parallel()

	srv := newTestPortal(t, "ABCD")
	defer srv.Close()

	solver := &mockSolver{solutions: []string{"ABCD"}}
	f := newFetcherForTest(solver)
	client, _ := f.newClient()
	s := &session{client: client, limiter: f.limiter, solver: solver}

	// baseURL is a const; construct request URLs directly against the test server.
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+pathAccessKey, nil)
	body, err := s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("loading access key page: %v", err)
	}

	fs, err := parseFormState(body)
	if err != nil {
		t.Fatalf("parsing form state: %v", err)
	}
	if !fs.valid() {
		t.Fatal("expected valid form state")
	}

	req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+pathCaptcha+"?t=123", nil)
	imgBody, err := s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("fetching captcha: %v", err)
	}
	if len(imgBody) == 0 {
		t.Error("expected non-empty captcha body")
	}

	form := buildForm(fs, map[string]string{
		fieldAccessKey: "29240112345678000190650010000001231234567890",
		fieldCaptcha:   "ABCD",
		fieldSubmit:    "Consultar",
	})
	req, _ = http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+pathAccessKey, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	body, err = s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("submitting access key: %v", err)
	}
	if err := checkErrors(body); err != nil {
		t.Fatalf("unexpected error in response: %v", err)
	}
	if !strings.Contains(string(body), "DANFE content") {
		t.Errorf("expected DANFE content, got: %s", truncate(string(body), 100))
	}
}

func TestSession_accessKeyFlow_captchaInvalid(t *testing.T) {
	t.Parallel()

	srv := newTestPortal(t, "CORRECT")
	defer srv.Close()

	f := newFetcherForTest(nil)
	client, _ := f.newClient()
	s := &session{client: client, limiter: f.limiter}

	// Submit with wrong captcha
	fs := &formState{viewState: "vs", eventValidation: "ev"}
	form := buildForm(fs, map[string]string{
		fieldAccessKey: "12345678901234567890123456789012345678901234",
		fieldCaptcha:   "WRONG",
		fieldSubmit:    "Consultar",
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+pathAccessKey, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	body, err := s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	err = checkErrors(body)
	if !errors.Is(err, sefaz.ErrCaptchaInvalid) {
		t.Errorf("expected ErrCaptchaInvalid, got: %v", err)
	}
}

func TestSession_pipeline(t *testing.T) {
	t.Parallel()

	srv := newTestPortal(t, "")
	defer srv.Close()

	f := newFetcherForTest(nil)
	client, _ := f.newClient()
	s := &session{client: client, limiter: f.limiter}

	// baseURL is a const; test pipeline steps individually against the test server.
	fs := &formState{viewState: "vs", eventValidation: "ev"}
	form := buildForm(fs, map[string]string{
		fieldViewTabs: "Visualizar em Abas",
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+pathDANFE, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	body, err := s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("navigateToTabs: %v", err)
	}
	if !strings.Contains(string(body), "nfe tab content") {
		t.Errorf("expected nfe tab, got: %s", truncate(string(body), 100))
	}

	req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+pathPrint+"?imprimir_nfe=1&print=true", nil)
	body, err = s.do(context.Background(), req)
	if err != nil {
		t.Fatalf("fetchPrintPage: %v", err)
	}
	if !strings.Contains(string(body), "print page content") {
		t.Errorf("expected print page content, got: %s", truncate(string(body), 100))
	}
}

func TestSession_contextCanceled(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	f := newFetcherForTest(nil)
	client, _ := f.newClient()
	s := &session{client: client, limiter: f.limiter}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	_, err := s.do(ctx, req)
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestValidateRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *sefaz.Request
		want error
	}{
		{
			name: "missing input",
			req:  &sefaz.Request{},
			want: sefaz.ErrMissingInput,
		},
		{
			name: "ambiguous input",
			req:  &sefaz.Request{AccessKey: "key", QRURL: "url"},
			want: sefaz.ErrAmbiguousInput,
		},
		{
			name: "access key only",
			req:  &sefaz.Request{AccessKey: "key"},
			want: nil,
		},
		{
			name: "qr url only",
			req:  &sefaz.Request{QRURL: "url"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validateRequest(tt.req)
			if !errors.Is(got, tt.want) {
				t.Errorf("validateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
