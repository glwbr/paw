package api

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/glwbr/paw/errs"
	apierrors "github.com/glwbr/paw/internal/api/errors"
)

type contextKey string

const ctxRequestID contextKey = "request_id"

type handler func(http.ResponseWriter, *http.Request) error

func handle(fn handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			writeError(w, r, err)
		}
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		err = apierrors.NotFound("receipt import not found")
	case errors.Is(err, ErrCaptchaNotPending):
		err = apierrors.Conflict("no captcha pending for this import")
	}

	reqID, _ := r.Context().Value(ctxRequestID).(string)

	var he *apierrors.HTTPError
	if errors.As(err, &he) {
		slog.Warn("api error", "route", r.Method+" "+r.Pattern, "status", he.Status, "request_id", reqID, "err", err)
		writeJSON(w, he.Status, errorBody(he.Msg, reqID, err))
		return
	}

	if msg := errs.PublicMessage(err); msg != "" {
		slog.Warn("api error", "route", r.Method+" "+r.Pattern, "request_id", reqID, "err", err)
		writeJSON(w, http.StatusBadRequest, errorBody(msg, reqID, err))
		return
	}

	slog.Error("api error", "route", r.Method+" "+r.Pattern, "request_id", reqID, "err", err)
	writeJSON(w, http.StatusInternalServerError, errorBody("internal error", reqID, err))
}

// errorBody builds the response payload. In dev mode it includes a "detail" field with the full error chain.
func errorBody(msg, reqID string, err error) map[string]string {
	body := map[string]string{
		"error":      msg,
		"request_id": reqID,
	}
	if isDev() {
		body["detail"] = err.Error()
	}
	return body
}

func isDev() bool {
	return os.Getenv("APP_ENV") == "dev" || os.Getenv("LOG_LEVEL") == "debug"
}
