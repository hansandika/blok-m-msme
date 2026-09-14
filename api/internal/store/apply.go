package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hansandika/blok-m-msme/api/internal/geofence"
	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/jackc/pgx/v5"
)

const communitySource = "Applied from a community suggestion. Hours and pins are approximate; confirm on-site."

const suggestionCols = `id, kind, place_id, name, category, notes, lat, lng, status, created_at, applied_at, applied_place_id`

// ApplyResult is returned after applying (or re-applying) a suggestion onto places.
type ApplyResult struct {
	Suggestion     *models.Suggestion `json:"suggestion"`
	Place          *models.Place      `json:"place,omitempty"`
	AlreadyApplied bool               `json:"alreadyApplied"`
}

/*
ApplySuggestion copies an approved suggestion onto live `places`.

Public contribute never writes places — only this admin path does.

new_place:
  - name from the suggestion (required)
  - slug from the name, uniqued with a short suffix on collision
  - id is a fresh UUID (not the seed SHA1 namespace)
  - category default "other"; price default "$$"; vibes default empty
  - lat/lng from the suggestion, else from notes, else the geofence center
    (pins must fall inside the Blok M box when provided)
  - neighborhood default "Blok M"; address default "<neighborhood>, Kebayoran Baru"
  - description defaults to the notes text
  - hours / open-late / extra fields: see notes overlay below
  - source_note records that this row came from a suggestion

correction:
  - requires a live place_id
  - structured suggestion fields (name, category, lat, lng) overwrite when set
  - notes overlay (key:value lines, case-insensitive keys):
      name, category, price, hours, neighborhood, address, description,
      source / sourceNote, vibe / vibes (comma list of known tags),
      openLate / open-late / larut (true/false), lat, lng
  - bare tokens: "open late" / "larut" set openLate; "not open late" / "closes early"
    clear it; a lone $ / $$ / $$$ sets price
  - if nothing structured matched, notes are appended as a community comment
    on description so the report is not discarded
  - source_note always gains an "applied from suggestion {id}" suffix

Rejected suggestions never mutate places. Applying twice is idempotent:
the same place is returned with alreadyApplied=true.
*/

func (s *Store) GetSuggestion(ctx context.Context, id string) (*models.Suggestion, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+suggestionCols+` FROM suggestions WHERE id = $1`, id)
	return scanSuggestion(row)
}

func (s *Store) SetSuggestionStatus(ctx context.Context, id, status string) (*models.Suggestion, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "rejected" {
		return nil, errf(400, "status must be approved or rejected")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	sug, err := scanSuggestion(tx.QueryRow(ctx, `SELECT `+suggestionCols+` FROM suggestions WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return nil, err
	}
	if sug == nil {
		return nil, errf(404, "suggestion not found")
	}
	if sug.AppliedAt != nil {
		return nil, errf(409, "applied suggestions cannot change status")
	}
	if sug.Status == status {
		return sug, tx.Commit(ctx)
	}

	sug, err = scanSuggestion(tx.QueryRow(ctx, `
UPDATE suggestions SET status = $2 WHERE id = $1
RETURNING `+suggestionCols, id, status))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return sug, nil
}

func (s *Store) ApplySuggestion(ctx context.Context, id string) (*ApplyResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	sug, err := scanSuggestion(tx.QueryRow(ctx, `SELECT `+suggestionCols+` FROM suggestions WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return nil, err
	}
	if sug == nil {
		return nil, errf(404, "suggestion not found")
	}
	if sug.Status == "rejected" {
		return nil, errf(409, "rejected suggestions cannot be applied")
	}
	if sug.AppliedAt != nil {
		place, err := getPlaceTx(ctx, tx, deref(sug.AppliedPlaceID))
		if err != nil {
			return nil, err
		}
		if place == nil && sug.PlaceID != nil {
			place, err = getPlaceTx(ctx, tx, *sug.PlaceID)
			if err != nil {
				return nil, err
			}
		}
		return &ApplyResult{Suggestion: sug, Place: place, AlreadyApplied: true}, tx.Commit(ctx)
	}
	if sug.Status != "approved" {
		return nil, errf(400, "approve the suggestion before applying")
	}

	var place *models.Place
	switch sug.Kind {
	case "new_place":
		place, err = insertPlaceFromSuggestion(ctx, tx, sug)
	case "correction":
		place, err = applyCorrectionTx(ctx, tx, sug)
	default:
		return nil, errf(400, "unknown suggestion kind")
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sug, err = scanSuggestion(tx.QueryRow(ctx, `
UPDATE suggestions
SET applied_at = $2, applied_place_id = $3
WHERE id = $1
RETURNING `+suggestionCols, sug.ID, now, place.ID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &ApplyResult{Suggestion: sug, Place: place, AlreadyApplied: false}, nil
}

func insertPlaceFromSuggestion(ctx context.Context, tx pgx.Tx, sug *models.Suggestion) (*models.Place, error) {
	patch := ParseNotes(sug.Notes)
	name := firstNonEmpty(deref(sug.Name), deref(patch.Name))
	if name == "" {
		return nil, errf(400, "name is required to apply a new place")
	}

	category := strings.ToLower(firstNonEmpty(deref(sug.Category), deref(patch.Category), "other"))
	if !allowedCategories[category] {
		return nil, errf(400, "invalid category")
	}
	price := "$$"
	if patch.Price != nil {
		price = *patch.Price
	}
	openLate := false
	if patch.OpenLate != nil {
		openLate = *patch.OpenLate
	}

	lat, lng, err := resolveCoords(sug.Lat, sug.Lng, patch.Lat, patch.Lng, true)
	if err != nil {
		return nil, err
	}

	neighborhood := firstNonEmpty(deref(patch.Neighborhood), "Blok M")
	address := firstNonEmpty(deref(patch.Address), neighborhood+", Kebayoran Baru")
	desc := firstNonEmpty(deref(patch.Description), strings.TrimSpace(sug.Notes))
	if desc == "" {
		desc = name + " — added from a community suggestion."
	}
	source := communitySource + " Suggestion " + sug.ID + "."
	if patch.SourceNote != nil {
		source = strings.TrimSpace(*patch.SourceNote)
	}

	slug, err := uniqueSlugTx(ctx, tx, name)
	if err != nil {
		return nil, err
	}

	p := models.Place{
		ID:           uuid.NewString(),
		Slug:         slug,
		Name:         name,
		Category:     category,
		Vibes:        mergeVibes(nil, patch.Vibes),
		Price:        price,
		OpenLate:     openLate,
		Lat:          lat,
		Lng:          lng,
		Neighborhood: neighborhood,
		Address:      address,
		Hours:        patch.Hours,
		Description:  desc,
		SourceNote:   source,
		Aliases:      []string{},
	}

	if err := insertPlaceTx(ctx, tx, p); err != nil {
		return nil, err
	}
	return getPlaceTx(ctx, tx, p.ID)
}

func applyCorrectionTx(ctx context.Context, tx pgx.Tx, sug *models.Suggestion) (*models.Place, error) {
	if sug.PlaceID == nil || strings.TrimSpace(*sug.PlaceID) == "" {
		return nil, errf(400, "correction is missing a place id")
	}
	place, err := getPlaceTx(ctx, tx, *sug.PlaceID)
	if err != nil {
		return nil, err
	}
	if place == nil {
		return nil, errf(404, "place not found for this correction")
	}

	patch := ParseNotes(sug.Notes)
	changed := false

	if name := firstNonEmpty(deref(patch.Name), deref(sug.Name)); name != "" && name != place.Name {
		place.Name = name
		changed = true
	}
	if cat := firstNonEmpty(deref(patch.Category), strings.ToLower(deref(sug.Category))); cat != "" {
		if !allowedCategories[cat] {
			return nil, errf(400, "invalid category")
		}
		if cat != place.Category {
			place.Category = cat
			changed = true
		}
	}
	if patch.Price != nil && *patch.Price != place.Price {
		place.Price = *patch.Price
		changed = true
	}
	if patch.OpenLate != nil && *patch.OpenLate != place.OpenLate {
		place.OpenLate = *patch.OpenLate
		changed = true
	}
	if patch.Neighborhood != nil && *patch.Neighborhood != place.Neighborhood {
		place.Neighborhood = *patch.Neighborhood
		changed = true
	}
	if patch.Address != nil && *patch.Address != place.Address {
		place.Address = *patch.Address
		changed = true
	}
	if patch.Hours != nil {
		if place.Hours == nil || *place.Hours != *patch.Hours {
			place.Hours = patch.Hours
			changed = true
		}
	}
	if patch.Description != nil && *patch.Description != place.Description {
		place.Description = *patch.Description
		changed = true
	}
	if len(patch.Vibes) > 0 {
		merged := mergeVibes(place.Vibes, patch.Vibes)
		if strings.Join(merged, ",") != strings.Join(place.Vibes, ",") {
			place.Vibes = merged
			changed = true
		}
	}

	lat, lng, err := resolveCoords(sug.Lat, sug.Lng, patch.Lat, patch.Lng, false)
	if err != nil {
		return nil, err
	}
	if (sug.Lat != nil && sug.Lng != nil) || (patch.Lat != nil && patch.Lng != nil) {
		if lat != place.Lat || lng != place.Lng {
			place.Lat = lat
			place.Lng = lng
			changed = true
		}
	}

	if !changed {
		note := strings.TrimSpace(sug.Notes)
		if note != "" && !strings.Contains(place.Description, note) {
			place.Description = strings.TrimSpace(place.Description + "\n\nCommunity note: " + note)
			changed = true
		}
	}

	suffix := "Applied from community suggestion " + sug.ID + "."
	if patch.SourceNote != nil {
		place.SourceNote = strings.TrimSpace(*patch.SourceNote)
	}
	if !strings.Contains(place.SourceNote, sug.ID) {
		place.SourceNote = strings.TrimSpace(place.SourceNote + " " + suffix)
	}

	if err := updatePlaceTx(ctx, tx, *place); err != nil {
		return nil, err
	}
	return getPlaceTx(ctx, tx, place.ID)
}

func resolveCoords(aLat, aLng, bLat, bLng *float64, defaultCenter bool) (float64, float64, error) {
	lat, lng := aLat, aLng
	if lat == nil {
		lat = bLat
	}
	if lng == nil {
		lng = bLng
	}
	if (lat == nil) != (lng == nil) {
		return 0, 0, errf(400, "lat and lng must both be set")
	}
	if lat == nil || lng == nil {
		if !defaultCenter {
			return 0, 0, nil
		}
		return geofence.CenterLat, geofence.CenterLng, nil
	}
	if !geofence.Contains(*lat, *lng) {
		return 0, 0, errf(400, "pin is outside the Blok M geofence")
	}
	return *lat, *lng, nil
}

func uniqueSlugTx(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	base := Slugify(name)
	slug := base
	for i := 0; i < 12; i++ {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM places WHERE slug = $1`, slug).Scan(&n); err != nil {
			return "", err
		}
		if n == 0 {
			return slug, nil
		}
		slug = base + "-" + uuid.NewString()[:6]
	}
	return base + "-" + uuid.NewString()[:8], nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func insertPlaceTx(ctx context.Context, tx pgx.Tx, p models.Place) error {
	_, err := tx.Exec(ctx, `
INSERT INTO places (
  id, slug, name, category, vibes, price, open_late, lat, lng,
  neighborhood, address, hours, description, source_note, aliases, updated_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,
  $10,$11,$12,$13,$14,$15, now()
)`,
		p.ID, p.Slug, p.Name, p.Category, p.Vibes, p.Price, p.OpenLate, p.Lat, p.Lng,
		p.Neighborhood, p.Address, p.Hours, p.Description, p.SourceNote, p.Aliases,
	)
	return err
}

func updatePlaceTx(ctx context.Context, tx pgx.Tx, p models.Place) error {
	_, err := tx.Exec(ctx, `
UPDATE places SET
  name = $2, category = $3, vibes = $4, price = $5, open_late = $6,
  lat = $7, lng = $8, neighborhood = $9, address = $10, hours = $11,
  description = $12, source_note = $13, aliases = $14, updated_at = now()
WHERE id = $1`,
		p.ID, p.Name, p.Category, p.Vibes, p.Price, p.OpenLate,
		p.Lat, p.Lng, p.Neighborhood, p.Address, p.Hours,
		p.Description, p.SourceNote, p.Aliases,
	)
	return err
}

func getPlaceTx(ctx context.Context, tx pgx.Tx, idOrSlug string) (*models.Place, error) {
	if strings.TrimSpace(idOrSlug) == "" {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `
SELECT id, slug, name, category, vibes, price, open_late, lat, lng,
       neighborhood, address, hours, description, source_note, aliases
FROM places
WHERE slug = $1 OR id::text = $1
LIMIT 1`, idOrSlug)
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

type suggestionScanner interface {
	Scan(dest ...any) error
}

func scanSuggestion(row suggestionScanner) (*models.Suggestion, error) {
	var out models.Suggestion
	var pid, appliedPID *string
	var appliedAt *time.Time
	err := row.Scan(
		&out.ID, &out.Kind, &pid, &out.Name, &out.Category, &out.Notes,
		&out.Lat, &out.Lng, &out.Status, &out.CreatedAt, &appliedAt, &appliedPID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	out.PlaceID = pid
	out.AppliedAt = appliedAt
	out.AppliedPlaceID = appliedPID
	return &out, nil
}
