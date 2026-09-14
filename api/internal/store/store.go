package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ListPlaces(ctx context.Context, f models.PlaceFilters) ([]models.Place, error) {
	q := `
SELECT id, slug, name, category, vibes, price, open_late, lat, lng,
       neighborhood, address, hours, description, source_note, aliases
FROM places
WHERE 1=1`
	args := []any{}
	n := 1

	if f.Query != "" {
		q += fmt.Sprintf(` AND (
			name ILIKE $%d OR neighborhood ILIKE $%d
			OR address ILIKE $%d OR category ILIKE $%d OR slug ILIKE $%d
			OR EXISTS (SELECT 1 FROM unnest(aliases) a WHERE a ILIKE $%d)
			OR EXISTS (SELECT 1 FROM unnest(vibes) v WHERE v ILIKE $%d)
		)`, n, n, n, n, n, n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.Category != "" {
		q += fmt.Sprintf(` AND category = $%d`, n)
		args = append(args, f.Category)
		n++
	}
	if f.Vibe != "" {
		q += fmt.Sprintf(` AND $%d = ANY(vibes)`, n)
		args = append(args, f.Vibe)
		n++
	}
	if f.Price != "" {
		q += fmt.Sprintf(` AND price = $%d`, n)
		args = append(args, f.Price)
		n++
	}
	if f.OpenLate != nil {
		q += fmt.Sprintf(` AND open_late = $%d`, n)
		args = append(args, *f.OpenLate)
		n++
	}
	if f.BBox != nil {
		q += fmt.Sprintf(` AND lng >= $%d AND lng <= $%d AND lat >= $%d AND lat <= $%d`, n, n+1, n+2, n+3)
		args = append(args, f.BBox.West, f.BBox.East, f.BBox.South, f.BBox.North)
		n += 4
	}

	order := ` ORDER BY name ASC`
	if f.NearLat != nil && f.NearLng != nil {
		q += fmt.Sprintf(` AND (
			6371000 * acos(LEAST(1.0, GREATEST(-1.0,
				cos(radians($%d)) * cos(radians(lat)) * cos(radians(lng) - radians($%d))
				+ sin(radians($%d)) * sin(radians(lat))
			)))
		) <= $%d`, n, n+1, n, n+2)
		radius := f.RadiusM
		if radius <= 0 {
			radius = 1200
		}
		args = append(args, *f.NearLat, *f.NearLng, radius)
		n += 3
		order = fmt.Sprintf(` ORDER BY (
			6371000 * acos(LEAST(1.0, GREATEST(-1.0,
				cos(radians($%d)) * cos(radians(lat)) * cos(radians(lng) - radians($%d))
				+ sin(radians($%d)) * sin(radians(lat))
			)))
		) ASC, name ASC`, n, n+1, n)
		args = append(args, *f.NearLat, *f.NearLng)
	}

	rows, err := s.pool.Query(ctx, q+order, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlaces(rows)
}

func (s *Store) GetPlace(ctx context.Context, idOrSlug string) (*models.Place, error) {
	const q = `
SELECT id, slug, name, category, vibes, price, open_late, lat, lng,
       neighborhood, address, hours, description, source_note, aliases
FROM places
WHERE slug = $1 OR id::text = $1
LIMIT 1`
	rows, err := s.pool.Query(ctx, q, idOrSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	places, err := scanPlaces(rows)
	if err != nil {
		return nil, err
	}
	if len(places) == 0 {
		return nil, nil
	}
	return &places[0], nil
}

func (s *Store) CreateSuggestion(ctx context.Context, in models.SuggestionInput) (*models.Suggestion, error) {
	const q = `
INSERT INTO suggestions (id, kind, place_id, name, category, notes, lat, lng)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING ` + suggestionCols

	var placeID any
	if strings.TrimSpace(in.PlaceID) != "" {
		placeID = in.PlaceID
	}
	var name any
	if strings.TrimSpace(in.Name) != "" {
		name = in.Name
	}
	var category any
	if strings.TrimSpace(in.Category) != "" {
		category = in.Category
	}

	return scanSuggestion(s.pool.QueryRow(ctx, q, uuid.NewString(), in.Kind, placeID, name, category, in.Notes, in.Lat, in.Lng))
}

func (s *Store) ListSuggestions(ctx context.Context, status string) ([]models.Suggestion, error) {
	q := `SELECT ` + suggestionCols + ` FROM suggestions`
	args := []any{}
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "":
	case "applied":
		q += ` WHERE applied_at IS NOT NULL`
	default:
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC LIMIT 200`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Suggestion{}
	for rows.Next() {
		item, err := scanSuggestion(rows)
		if err != nil {
			return nil, err
		}
		if item != nil {
			out = append(out, *item)
		}
	}
	return out, rows.Err()
}

func scanPlaces(rows pgx.Rows) ([]models.Place, error) {
	out := []models.Place{}
	for rows.Next() {
		var p models.Place
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Category, &p.Vibes, &p.Price, &p.OpenLate,
			&p.Lat, &p.Lng, &p.Neighborhood, &p.Address, &p.Hours, &p.Description,
			&p.SourceNote, &p.Aliases,
		); err != nil {
			return nil, err
		}
		if p.Vibes == nil {
			p.Vibes = []string{}
		}
		if p.Aliases == nil {
			p.Aliases = []string{}
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func ParseBBox(s string) (*models.BBox, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("bbox must be west,south,east,north")
	}
	nums := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("bbox: invalid number")
		}
		nums[i] = v
	}
	return &models.BBox{West: nums[0], South: nums[1], East: nums[2], North: nums[3]}, nil
}

func ParseBool(s string) (*bool, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y":
		t := true
		return &t, nil
	case "0", "false", "no", "n":
		t := false
		return &t, nil
	default:
		return nil, fmt.Errorf("invalid boolean")
	}
}
