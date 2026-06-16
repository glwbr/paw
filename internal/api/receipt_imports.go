package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/glwbr/paw/errs"
	apierrors "github.com/glwbr/paw/internal/api/errors"
	"github.com/glwbr/paw/internal/db"
	"github.com/glwbr/paw/internal/store"
	"github.com/glwbr/paw/nfce"
	"github.com/glwbr/paw/sefaz"
	"github.com/glwbr/paw/sefaz/captcha"
)

const importTimeout = 5 * time.Minute

// Status represents the current state of a receipt import.
type Status string

const (
	StatusPending        Status = "pending"
	StatusFetching       Status = "fetching"
	StatusWaitingCaptcha Status = "waiting_captcha"
	StatusProcessing     Status = "processing"
	StatusCompleted      Status = "completed"
	StatusFailed         Status = "failed"
)

// Terminal reports whether the status is a final state.
func (s Status) Terminal() bool { return s == StatusCompleted || s == StatusFailed }

// Sentinel errors for flow control. Checked with errors.Is; mapped to HTTP
// responses in respond.go.
var (
	ErrNotFound          = errors.New("receipt import not found")
	ErrCaptchaNotPending = errors.New("no captcha pending for this import")
)

// fetcherFactory creates a Fetcher, optionally configured with a captcha solver.
type fetcherFactory func(captcha.Solver) sefaz.Fetcher

// ReceiptImport tracks the lifecycle of one NFC-e import from SEFAZ. It is
// its own captcha.Solver: the fetcher's blocking Solve call bridges to the
// HTTP captcha endpoint through answerCh, with no separate solver type.
type ReceiptImport struct {
	ID        string
	AccessKey string
	QRURL     string
	ReceiptID *int64
	CreatedAt time.Time
	UpdatedAt time.Time

	mu        sync.Mutex
	status    Status
	errMsg    string
	captcha   []byte
	captchaCT string
	answerCh  chan string
}

func newReceiptImport(accessKey, qrURL string) *ReceiptImport {
	now := time.Now()
	return &ReceiptImport{
		ID:        newID(),
		AccessKey: accessKey,
		QRURL:     qrURL,
		CreatedAt: now,
		UpdatedAt: now,
		status:    StatusPending,
	}
}

// Status returns the current lifecycle state under the import's lock.
func (r *ReceiptImport) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

func (r *ReceiptImport) setStatus(s Status) {
	r.mu.Lock()
	r.status = s
	r.UpdatedAt = time.Now()
	r.mu.Unlock()
}

// fail stores a client-safe message on the import and logs the raw cause
// server-side. clientMsg is what reaches the API; internal is logged only.
func (r *ReceiptImport) fail(clientMsg string, start time.Time, internal any) {
	r.mu.Lock()
	r.status = StatusFailed
	r.errMsg = clientMsg
	r.UpdatedAt = time.Now()
	r.mu.Unlock()

	attrs := []any{
		"id", r.ID,
		"client_msg", clientMsg,
		"duration_ms", time.Since(start).Milliseconds(),
	}
	if internal != nil {
		attrs = append(attrs, "internal", internal)
	}
	slog.Error("import failed", attrs...)
}

// captchaImage returns the current challenge image and its content-type.
// Returns (nil, "") when no challenge is pending.
func (r *ReceiptImport) captchaImage() ([]byte, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.captcha) == 0 {
		return nil, ""
	}
	img := make([]byte, len(r.captcha))
	copy(img, r.captcha)
	return img, r.captchaCT
}

// Solve implements captcha.Solver. Called from the sefaz fetcher goroutine;
// publishes the challenge image on the import, flips status to waiting_captcha,
// and blocks until submitCaptcha delivers an answer or ctx is cancelled.
func (r *ReceiptImport) Solve(ctx context.Context, ch captcha.Challenge) (captcha.Solution, error) {
	r.mu.Lock()
	r.captcha = ch.Image
	r.captchaCT = ch.ContentType
	r.status = StatusWaitingCaptcha
	r.answerCh = make(chan string, 1)
	answerCh := r.answerCh
	r.UpdatedAt = time.Now()
	r.mu.Unlock()

	slog.Info("captcha required", "id", r.ID)

	select {
	case ans := <-answerCh:
		return captcha.Solution{Text: ans, ChallengeID: ch.ID}, nil
	case <-ctx.Done():
		return captcha.Solution{}, ctx.Err()
	}
}

// submitCaptcha delivers a captcha answer to a Solve call blocked on answerCh.
// Returns ErrCaptchaNotPending if no Solve is currently waiting.
func (r *ReceiptImport) submitCaptcha(answer string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusWaitingCaptcha || r.answerCh == nil {
		return ErrCaptchaNotPending
	}
	r.answerCh <- answer
	r.answerCh = nil
	return nil
}

// run is the linear import pipeline: fetch -> parse -> save. Started as a
// goroutine from Server.submitImport. 5-minute timeout covers the whole
// pipeline, including time the user spends typing the captcha answer.
func (r *ReceiptImport) run(newFetcher fetcherFactory, q *db.Queries) {
	ctx, cancel := context.WithTimeout(context.Background(), importTimeout)
	defer cancel()

	start := time.Now()
	r.setStatus(StatusFetching)

	var solver captcha.Solver
	if r.AccessKey != "" {
		solver = r
	}

	res, err := newFetcher(solver).Fetch(ctx, &sefaz.Request{
		AccessKey: r.AccessKey,
		QRURL:     r.QRURL,
	})
	if err != nil {
		r.fail(fetchClientMessage(err), start, err)
		return
	}

	r.setStatus(StatusProcessing)
	parsed := nfce.Parse(nfce.Page{Name: "import", Content: res.Page})
	if parsed.HasErrors() {
		r.fail("failed to parse receipt", start, parsed.Errors)
		return
	}

	if _, err := store.SaveReceipt(ctx, q, *parsed.Receipt); err != nil {
		r.fail("failed to save receipt", start, err)
		return
	}

	r.mu.Lock()
	r.status = StatusCompleted
	r.AccessKey = parsed.Receipt.AccessKey
	r.UpdatedAt = time.Now()
	r.mu.Unlock()

	slog.Info("import completed",
		"id", r.ID,
		"access_key", parsed.Receipt.AccessKey,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// fetchClientMessage extracts a client-safe message from a sefaz fetcher
// error. Known portal conditions surface their sanitised messages; anything
// else becomes a generic string and the raw cause is logged by fail.
func fetchClientMessage(err error) string {
	if msg := errs.PublicMessage(err); msg != "" {
		return msg
	}
	return "failed to fetch receipt from portal"
}

// MarshalJSON serialises the import for API responses. Follows the LRO
// convention of exposing a `done` boolean so clients have a single field to
// stop polling on.
func (r *ReceiptImport) MarshalJSON() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var captchaURL string
	if r.status == StatusWaitingCaptcha {
		captchaURL = "/receipts/imports/" + r.ID + "/captcha"
	}

	return json.Marshal(struct {
		ID         string    `json:"id"`
		Status     Status    `json:"status"`
		Done       bool      `json:"done"`
		AccessKey  string    `json:"access_key,omitempty"`
		QRURL      string    `json:"qr_url,omitempty"`
		ReceiptID  *int64    `json:"receipt_id,omitempty"`
		CaptchaURL string    `json:"captcha_url,omitempty"`
		Error      string    `json:"error,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}{
		ID:         r.ID,
		Status:     r.status,
		Done:       r.status.Terminal(),
		AccessKey:  r.AccessKey,
		QRURL:      r.QRURL,
		ReceiptID:  r.ReceiptID,
		CaptchaURL: captchaURL,
		Error:      r.errMsg,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	})
}

// submitImport creates a receipt import and starts its run goroutine. Returns
// the existing non-terminal import when the same access key is already in
// flight, so retrying is idempotent at the access-key level.
func (s *Server) submitImport(accessKey, qrURL string) *ReceiptImport {
	s.mu.Lock()
	defer s.mu.Unlock()

	if accessKey != "" {
		if existing := s.findActiveByAccessKeyLocked(accessKey); existing != nil {
			return existing
		}
	}

	ri := newReceiptImport(accessKey, qrURL)
	s.imports[ri.ID] = ri
	go ri.run(s.newFetcher, s.queries)
	return ri
}

func (s *Server) getImportByID(id string) (*ReceiptImport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ri, ok := s.imports[id]
	if !ok {
		return nil, ErrNotFound
	}
	return ri, nil
}

func (s *Server) findActiveByAccessKeyLocked(accessKey string) *ReceiptImport {
	for _, ri := range s.imports {
		if ri.AccessKey != accessKey {
			continue
		}
		if !ri.Status().Terminal() {
			return ri
		}
	}
	return nil
}

func (s *Server) createImport(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		AccessKey string `json:"access_key"`
		QRURL     string `json:"qr_url"`
	}
	if err := decode(r, &req); err != nil {
		return apierrors.BadRequest("invalid request body", err)
	}
	if (req.AccessKey == "") == (req.QRURL == "") {
		return apierrors.BadRequest("exactly one of access_key or qr_url is required", nil)
	}

	ri := s.submitImport(req.AccessKey, req.QRURL)

	w.Header().Set("Operation-Location", "/receipts/imports/"+ri.ID)
	encode(w, http.StatusAccepted, ri)
	return nil
}

func (s *Server) getImport(w http.ResponseWriter, r *http.Request) error {
	ri, err := s.getImportByID(r.PathValue("id"))
	if err != nil {
		return err
	}

	encode(w, http.StatusOK, ri)
	return nil
}

func (s *Server) getCaptchaImage(w http.ResponseWriter, r *http.Request) error {
	ri, err := s.getImportByID(r.PathValue("id"))
	if err != nil {
		return err
	}

	img, ct := ri.captchaImage()
	if len(img) == 0 {
		return apierrors.NotFound("no captcha available")
	}

	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(img)
	return nil
}

func (s *Server) postCaptchaAnswer(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Answer string `json:"answer"`
	}
	if err := decode(r, &req); err != nil {
		return apierrors.BadRequest("invalid request body", err)
	}
	if req.Answer == "" {
		return apierrors.BadRequest("answer is required", nil)
	}

	ri, err := s.getImportByID(r.PathValue("id"))
	if err != nil {
		return err
	}
	if err := ri.submitCaptcha(req.Answer); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
