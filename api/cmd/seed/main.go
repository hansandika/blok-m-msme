package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/hansandika/blok-m-msme/api/internal/config"
	"github.com/hansandika/blok-m-msme/api/internal/fsdata"
	"github.com/hansandika/blok-m-msme/api/internal/seed"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := fsdata.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	raw := fsdata.PlacesJSON
	if path := os.Getenv("SEED_FILE"); path != "" {
		raw, err = os.ReadFile(path)
		if err != nil {
			log.Fatalf("read SEED_FILE: %v", err)
		}
	}

	n, err := seed.Upsert(ctx, pool, raw)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("upserted %d places", n)
}
