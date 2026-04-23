// Package ba implements the SEFAZ NFC-e fetcher for Bahia.
package ba

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/time/rate"

	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

const (
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	acceptHTML = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
	acceptLang = "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7"
)

// Fetcher implements sefaz.Fetcher for the Bahia SEFAZ NFC-e portal.
type Fetcher struct {
	solver  captcha.Solver
	limiter *rate.Limiter
}

// Option configures a Fetcher.
type Option func(*Fetcher)

// WithSolver sets the captcha solver used for access-key lookups.
func WithSolver(s captcha.Solver) Option {
	return func(f *Fetcher) { f.solver = s }
}

// New creates a BA Fetcher with default rate limiting.
func New(opts ...Option) *Fetcher {
	f := &Fetcher{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Fetch retrieves all pages from the BA SEFAZ portal (nfe, emitente, produtos, cobranca, totais).
func (f *Fetcher) Fetch(ctx context.Context, req *sefaz.Request) (*sefaz.Result, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	client, err := f.newClient()
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client: %w", err)
	}

	s := &session{
		client:  client,
		limiter: f.limiter,
		solver:  f.solver,
	}

	var danfeHTML []byte
	if req.QRURL != "" {
		danfeHTML, err = s.fetchQR(ctx, req.QRURL)
	} else {
		danfeHTML, err = s.fetchAccessKey(ctx, req.AccessKey)
	}
	if err != nil {
		return nil, err
	}

	return s.pipeline(ctx, danfeHTML)
}

func (f *Fetcher) newClient() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // BA portal uses self-signed cert
		},
	}, nil
}

func validateRequest(req *sefaz.Request) error {
	if req.AccessKey == "" && req.QRURL == "" {
		return sefaz.ErrMissingInput
	}
	if req.AccessKey != "" && req.QRURL != "" {
		return sefaz.ErrAmbiguousInput
	}
	return nil
}

// session holds per-Fetch state (fresh client, shared limiter/solver).
type session struct {
	client    *http.Client
	limiter   *rate.Limiter
	solver    captcha.Solver
	danfeBase string
}

// do executes a rate-limited HTTP request with browser headers and returns the response body.
func (s *session) do(ctx context.Context, req *http.Request) ([]byte, error) {
	body, _, err := s.doFull(ctx, req)
	return body, err
}

// doFull is like do but also returns the final *http.Response after redirects.
func (s *session) doFull(ctx context.Context, req *http.Request) ([]byte, *http.Response, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, nil, fmt.Errorf("rate limiter: %w", err)
	}

	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", acceptHTML)
	req.Header.Set("Accept-Language", acceptLang)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request to %s: %w", req.URL.Path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response from %s: %w", req.URL.Path, err)
	}

	if resp.StatusCode >= 400 {
		return nil, nil, &sefaz.PortalError{
			StatusCode: resp.StatusCode,
			Body:       truncate(string(body), 200),
		}
	}

	return body, resp, nil
}

func (s *session) fetchQR(ctx context.Context, qrURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, qrURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building QR request: %w", err)
	}

	body, resp, err := s.doFull(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetching QR URL: %w", err)
	}

	if err := checkErrors(body); err != nil {
		return nil, err
	}

	// Record the scheme+host of the final redirect destination so that
	// navigateToTabs posts to the same origin as the DANFE session.
	s.danfeBase = resp.Request.URL.Scheme + "://" + resp.Request.URL.Host
	return body, nil
}

func (s *session) fetchAccessKey(ctx context.Context, accessKey string) ([]byte, error) {
	if s.solver == nil {
		return nil, sefaz.ErrNoCaptchaSolver
	}

	fs, err := s.loadAccessKeyPage(ctx)
	if err != nil {
		return nil, err
	}

	for attempt := range maxCaptchaRetries {
		challenge, err := s.fetchCaptchaImage(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching captcha (attempt %d): %w", attempt+1, err)
		}

		solution, err := s.solver.Solve(ctx, challenge)
		if err != nil {
			return nil, &sefaz.CaptchaError{Err: err}
		}

		body, err := s.submitAccessKey(ctx, fs, accessKey, solution.Text)
		if errors.Is(err, sefaz.ErrCaptchaInvalid) {
			continue
		}
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, &sefaz.CaptchaError{
		Err: fmt.Errorf("failed after %d attempts: %w", maxCaptchaRetries, sefaz.ErrCaptchaInvalid),
	}
}

func (s *session) loadAccessKeyPage(ctx context.Context) (*formState, error) {
	u := baseURL + pathAccessKey
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building access key page request: %w", err)
	}

	body, err := s.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("loading access key page: %w", err)
	}

	fs, err := parseFormState(body)
	if err != nil {
		return nil, fmt.Errorf("parsing access key page: %w", err)
	}
	if !fs.valid() {
		return nil, fmt.Errorf("access key page: missing form state")
	}

	return fs, nil
}

func (s *session) fetchCaptchaImage(ctx context.Context) (captcha.Challenge, error) {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	u := baseURL + pathCaptcha + "?t=" + ts

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return captcha.Challenge{}, fmt.Errorf("building captcha request: %w", err)
	}
	req.Header.Set("Referer", baseURL+pathAccessKey)

	if err := s.limiter.Wait(ctx); err != nil {
		return captcha.Challenge{}, fmt.Errorf("rate limiter: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", acceptLang)

	resp, err := s.client.Do(req)
	if err != nil {
		return captcha.Challenge{}, fmt.Errorf("fetching captcha image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return captcha.Challenge{}, fmt.Errorf("captcha image returned HTTP %d", resp.StatusCode)
	}

	img, err := io.ReadAll(resp.Body)
	if err != nil {
		return captcha.Challenge{}, fmt.Errorf("reading captcha image: %w", err)
	}

	// SEFAZ BA returns image bytes with a wrong Content-Type (text/html),
	// so we detect the real type from the content itself.
	ct := http.DetectContentType(img)
	if !strings.HasPrefix(ct, "image/") {
		return captcha.Challenge{}, fmt.Errorf("captcha response is not an image: detected %s", ct)
	}

	return captcha.Challenge{
		ID:          ts,
		Image:       img,
		ContentType: ct,
	}, nil
}

func (s *session) submitAccessKey(ctx context.Context, fs *formState, accessKey, captchaText string) ([]byte, error) {
	form := buildForm(fs, map[string]string{
		fieldAccessKey: accessKey,
		fieldCaptcha:   captchaText,
		fieldSubmit:    "Consultar",
	})

	u := baseURL + pathAccessKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building access key submit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", u)

	body, err := s.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("submitting access key: %w", err)
	}

	if err := checkErrors(body); err != nil {
		return nil, err
	}

	s.danfeBase = baseURL // access key path always uses HTTPS
	return body, nil
}

func (s *session) pipeline(ctx context.Context, danfeHTML []byte) (*sefaz.Result, error) {
	fs, err := parseFormState(danfeHTML)
	if err != nil {
		return nil, fmt.Errorf("parsing DANFE form state: %w", err)
	}

	if err = s.navigateToTabs(ctx, fs); err != nil {
		return nil, fmt.Errorf("navigating to tabs: %w", err)
	}

	printHTML, err := s.fetchPrintPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching print page: %w", err)
	}

	return &sefaz.Result{Page: printHTML}, nil
}

func (s *session) fetchPrintPage(ctx context.Context) ([]byte, error) {
	u := s.danfeBase + pathPrint + "?imprimir_nfe=1&print=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building print page request: %w", err)
	}
	req.Header.Set("Referer", s.danfeBase+pathDANFE)

	body, err := s.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetching print page: %w", err)
	}

	if err := checkErrors(body); err != nil {
		return nil, err
	}

	return body, nil
}

func (s *session) navigateToTabs(ctx context.Context, fs *formState) error {
	form := buildForm(fs, map[string]string{
		fieldViewTabs: "Visualizar em Abas",
	})

	u := s.danfeBase + pathDANFE
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("building tabs request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", u)

	body, err := s.do(ctx, req)
	if err != nil {
		return err
	}

	return checkErrors(body)
}

// checkErrors scans HTML response body for portal error messages.
func checkErrors(html []byte) error {
	lower := strings.ToLower(string(html))

	patterns := []struct {
		text string
		err  error
	}{
		{"captcha inválido", sefaz.ErrCaptchaInvalid},
		{"código incorreto tente novamente", sefaz.ErrCaptchaInvalid},
		{"chave de acesso inválida", &sefaz.InvalidAccessKeyError{}},
		{"nfc-e não encontrada", &sefaz.InvoiceNotFoundError{}},
		{"sessão expirada", &sefaz.SessionExpiredError{}},
		{"ocorreu um erro", &sefaz.PortalError{}},
		{"object reference not set", &sefaz.PortalError{}},
	}

	for _, p := range patterns {
		if strings.Contains(lower, p.text) {
			return p.err
		}
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
