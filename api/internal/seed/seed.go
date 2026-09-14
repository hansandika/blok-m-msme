package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultSource = "Hand-curated seed for the Blok M demo. Hours and pins are approximate — confirm on-site. Not live Google Places data."

var ns = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func Upsert(ctx context.Context, pool *pgxpool.Pool, raw []byte) (int, error) {
	var seeds []models.PlaceSeed
	if err := json.Unmarshal(raw, &seeds); err != nil {
		return 0, fmt.Errorf("parse places.json: %w", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	const q = `
INSERT INTO places (
  id, slug, name, category, vibes, price, open_late, lat, lng,
  neighborhood, address, hours, description, source_note, aliases, updated_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,
  $10,$11,$12,$13,$14,$15, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  category = EXCLUDED.category,
  vibes = EXCLUDED.vibes,
  price = EXCLUDED.price,
  open_late = EXCLUDED.open_late,
  lat = EXCLUDED.lat,
  lng = EXCLUDED.lng,
  neighborhood = EXCLUDED.neighborhood,
  address = EXCLUDED.address,
  hours = EXCLUDED.hours,
  description = EXCLUDED.description,
  source_note = EXCLUDED.source_note,
  aliases = EXCLUDED.aliases,
  updated_at = now()`

	for i, p := range seeds {
		if err := validate(p); err != nil {
			return 0, fmt.Errorf("place %d (%s): %w", i, p.Slug, err)
		}
		id := uuid.NewSHA1(ns, []byte("blokm:"+p.Slug)).String()
		source := p.SourceNote
		if strings.TrimSpace(source) == "" {
			source = defaultSource
		}
		hours := p.Hours
		var hoursArg any
		if strings.TrimSpace(hours) == "" {
			hoursArg = nil
		} else {
			hoursArg = hours
		}
		if _, err := tx.Exec(ctx, q,
			id, p.Slug, p.Name, p.Category, p.Vibes, p.Price, p.OpenLate, p.Lat, p.Lng,
			p.Neighborhood, p.Address, hoursArg, p.Description, source, p.Aliases,
		); err != nil {
			return 0, fmt.Errorf("upsert %s: %w", p.Slug, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(seeds), nil
}

func Count(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var n int
	err := pool.QueryRow(ctx, `SELECT count(*) FROM places`).Scan(&n)
	return n, err
}

func validate(p models.PlaceSeed) error {
	if strings.TrimSpace(p.Slug) == "" || strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("slug and name required")
	}
	switch p.Category {
	case "japanese", "street", "cafe", "retail", "other":
	default:
		return fmt.Errorf("invalid category %q", p.Category)
	}
	switch p.Price {
	case "$", "$$", "$$$":
	default:
		return fmt.Errorf("invalid price %q", p.Price)
	}
	if p.Lat == 0 || p.Lng == 0 {
		return fmt.Errorf("lat/lng required")
	}
	return nil
}