// Package handler exposes the Blok M Lokal API as a Vercel Go serverless function.
// vercel.json rewrites every path to /api, so the chi router sees the original URL.
package handler

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hansandika/blok-m-msme/api/internal/httpapi"
	"github.com/hansandika/blok-m-msme/api/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once    sync.Once
	app     http.Handler
	initErr error
)

func setup() {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		initErr = err
		return
	}
	// Serverless: tiny pool, simple protocol so it also works behind PgBouncer/Supavisor.
	cfg.MaxConns = 2
	cfg.MaxConnIdleTime = 30 * time.Second
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		initErr = err
		return
	}
	app = httpapi.New(store.New(pool), os.Getenv("ADMIN_TOKEN"))
}

// Handler is the Vercel entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(setup)
	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"database unavailable"}`))
		return
	}
	app.ServeHTTP(w, r)
}
