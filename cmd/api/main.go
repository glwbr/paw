package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/glwbr/paw/internal/api"
	"github.com/glwbr/paw/internal/db"
	_ "github.com/glwbr/paw/nfce/sefaz/ba"
	"github.com/glwbr/paw/sefaz"
	sefazba "github.com/glwbr/paw/sefaz/ba"
	"github.com/glwbr/paw/sefaz/captcha"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(os.Getenv("LOG_LEVEL")),
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})).With("service", "paw-api"))

	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("db pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	newFetcher := func(solver captcha.Solver) sefaz.Fetcher {
		if solver != nil {
			return sefazba.New(sefazba.WithSolver(solver))
		}
		return sefazba.New()
	}

	srv := api.NewServer(q, newFetcher)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	var handler http.Handler = srv
	if origin := os.Getenv("CORS_ORIGIN"); origin != "" {
		handler = api.CORS(origin, srv)
	}

	slog.Info("server started", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server error", "error", err)
	}
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
