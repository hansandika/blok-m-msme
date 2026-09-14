package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hansandika/blok-m-msme/api/internal/config"
	"github.com/hansandika/blok-m-msme/api/internal/fsdata"
	"github.com/hansandika/blok-m-msme/api/internal/httpapi"
	"github.com/hansandika/blok-m-msme/api/internal/seed"
	"github.com/hansandika/blok-m-msme/api/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	if err := fsdata.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if cfg.AutoSeed {
		n, err := seed.Count(ctx, pool)
		if err != nil {
			log.Fatalf("seed count: %v", err)
		}
		if n == 0 {
			loaded, err := seed.Upsert(ctx, pool, fsdata.PlacesJSON)
			if err != nil {
				log.Fatalf("auto-seed: %v", err)
			}
			log.Printf("auto-seeded %d places", loaded)
		} else {
			log.Printf("places already present (%d); skip auto-seed", n)
		}
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(store.New(pool), cfg.AdminToken),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("blokm api listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
