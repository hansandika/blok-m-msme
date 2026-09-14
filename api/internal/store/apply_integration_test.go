package store

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hansandika/blok-m-msme/api/internal/fsdata"
	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		url = "postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("no postgres: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres not reachable: %v", err)
	}
	if err := fsdata.Migrate(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(pool.Close)
	return New(pool)
}

func TestApproveRejectApplyNewPlace(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()

	lat, lng := -6.2442, 106.7995
	sug, err := st.CreateSuggestion(ctx, models.SuggestionInput{
		Kind:     "new_place",
		Name:     "M1 Test Kopi Corner",
		Category: "cafe",
		Notes:    "hours: 08:00–22:00\nopen-late: false\nneighborhood: Melawai\nA quiet test stall for moderation tests.",
		Lat:      &lat,
		Lng:      &lng,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1`, sug.ID)
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM places WHERE slug LIKE 'm1-test-kopi-corner%'`)
	})

	if _, err := st.ApplySuggestion(ctx, sug.ID); err == nil {
		t.Fatal("pending apply should fail")
	}

	approved, err := st.SetSuggestionStatus(ctx, sug.ID, "approved")
	if err != nil || approved.Status != "approved" {
		t.Fatalf("approve: %#v %v", approved, err)
	}
	again, err := st.SetSuggestionStatus(ctx, sug.ID, "approved")
	if err != nil || again.Status != "approved" {
		t.Fatalf("idempotent approve: %v", err)
	}

	res, err := st.ApplySuggestion(ctx, sug.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.AlreadyApplied || res.Place == nil {
		t.Fatalf("first apply: %#v", res)
	}
	if res.Place.Name != "M1 Test Kopi Corner" || res.Place.Category != "cafe" {
		t.Fatalf("place %#v", res.Place)
	}
	if res.Place.Neighborhood != "Melawai" {
		t.Fatalf("neighborhood %s", res.Place.Neighborhood)
	}
	got, err := st.GetPlace(ctx, res.Place.Slug)
	if err != nil || got == nil {
		t.Fatalf("get applied place: %v", err)
	}

	dup, err := st.ApplySuggestion(ctx, sug.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !dup.AlreadyApplied || dup.Place == nil || dup.Place.ID != res.Place.ID {
		t.Fatalf("second apply should be idempotent: %#v", dup)
	}
	if _, err := st.SetSuggestionStatus(ctx, sug.ID, "rejected"); err == nil {
		t.Fatal("applied suggestion should not change status")
	}
}

func TestRejectDoesNotMutatePlaces(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()

	before, err := st.ListPlaces(ctx, models.PlaceFilters{})
	if err != nil {
		t.Fatal(err)
	}

	sug, err := st.CreateSuggestion(ctx, models.SuggestionInput{
		Kind:  "new_place",
		Name:  "M1 Rejected Ghost Stall",
		Notes: "Should never land on the map.",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1`, sug.ID)
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM places WHERE slug LIKE 'm1-rejected-ghost-stall%'`)
	})

	if _, err := st.SetSuggestionStatus(ctx, sug.ID, "rejected"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ApplySuggestion(ctx, sug.ID); err == nil {
		t.Fatal("rejected apply should fail")
	}

	after, err := st.ListPlaces(ctx, models.PlaceFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("place count changed: %d -> %d", len(before), len(after))
	}
	place, err := st.GetPlace(ctx, "m1-rejected-ghost-stall")
	if err != nil {
		t.Fatal(err)
	}
	if place != nil {
		t.Fatal("rejected suggestion wrote a place")
	}
}

func TestApplyCorrection(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()

	seed, err := st.CreateSuggestion(ctx, models.SuggestionInput{
		Kind:     "new_place",
		Name:     "M1 Correction Target",
		Category: "other",
		Notes:    "hours: 10:00–18:00\nneighborhood: Melawai",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.SetSuggestionStatus(ctx, seed.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	created, err := st.ApplySuggestion(ctx, seed.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1 OR id = $2`, seed.ID, "")
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE applied_place_id = $1`, created.Place.ID)
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1`, seed.ID)
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM places WHERE id = $1`, created.Place.ID)
	})

	corr, err := st.CreateSuggestion(ctx, models.SuggestionInput{
		Kind:    "correction",
		PlaceID: created.Place.ID,
		Name:    created.Place.Name,
		Notes:   "hours: 17:00–01:00\nopen-late: true\nprice: $$\nvibes: quiet",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1`, corr.ID)
	})
	if _, err := st.SetSuggestionStatus(ctx, corr.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	res, err := st.ApplySuggestion(ctx, corr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.Place == nil || res.Place.Hours == nil || *res.Place.Hours != "17:00–01:00" {
		t.Fatalf("hours not applied: %#v", res.Place)
	}
	if !res.Place.OpenLate {
		t.Fatal("openLate not applied")
	}
	if res.Place.Price != "$$" {
		t.Fatalf("price %s", res.Place.Price)
	}
}

func TestApplyOutsideGeofence(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	lat, lng := -6.1000, 106.8000
	sug, err := st.CreateSuggestion(ctx, models.SuggestionInput{
		Kind:  "new_place",
		Name:  "M1 Far Away",
		Notes: "Outside the box.",
		Lat:   &lat,
		Lng:   &lng,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM suggestions WHERE id = $1`, sug.ID)
	})
	if _, err := st.SetSuggestionStatus(ctx, sug.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	_, err = st.ApplySuggestion(ctx, sug.ID)
	if err == nil {
		t.Fatal("expected geofence rejection")
	}
	se, ok := err.(*Error)
	if !ok || se.Code != 400 {
		t.Fatalf("want 400 geofence, got %#v", err)
	}
}
