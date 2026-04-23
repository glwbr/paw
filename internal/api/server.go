// Package api implements the HTTP REST API for NFC-e receipt ingestion.
package api

import (
	"net/http"
	"sync"

	"github.com/glwbr/paw/internal/db"
)

// Server handles HTTP API requests. Owns the in-memory receipt-import state
// directly — no Manager, no Store, no Broadcaster interfaces.
type Server struct {
	mux        *http.ServeMux
	queries    *db.Queries
	newFetcher fetcherFactory

	mu      sync.Mutex
	imports map[string]*ReceiptImport
}

func NewServer(q *db.Queries, f fetcherFactory) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		queries:    q,
		newFetcher: f,
		imports:    make(map[string]*ReceiptImport),
	}

	s.mux.HandleFunc("POST /receipts/imports", handle(s.createImport))
	s.mux.HandleFunc("GET /receipts/imports/{id}", handle(s.getImport))
	s.mux.HandleFunc("GET /receipts/imports/{id}/captcha", handle(s.getCaptchaImage))
	s.mux.HandleFunc("POST /receipts/imports/{id}/captcha", handle(s.postCaptchaAnswer))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID(s.mux).ServeHTTP(w, r)
}
