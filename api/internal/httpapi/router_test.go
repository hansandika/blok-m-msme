package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/hansandika/blok-m-msme/api/internal/store"
)

type stub struct {
	suggestions []models.Suggestion
	places      []models.Place
	created     *models.Suggestion
	statusID    string
	statusVal   string
	statusOut   *models.Suggestion
	statusErr   error
	applyID     string
	applyOut    *store.ApplyResult
	applyErr    error
}

func (s *stub) ListPlaces(context.Context, models.PlaceFilters) ([]models.Place, error) {
	return s.places, nil
}
func (s *stub) GetPlace(_ context.Context, id string) (*models.Place, error) {
	for i := range s.places {
		if s.places[i].ID == id || s.places[i].Slug == id {
			return &s.places[i], nil
		}
	}
	return nil, nil
}
func (s *stub) CreateSuggestion(_ context.Context, in models.SuggestionInput) (*models.Suggestion, error) {
	out := models.Suggestion{ID: "sug-1", Kind: in.Kind, Notes: in.Notes, Status: "pending"}
	if in.Name != "" {
		n := in.Name
		out.Name = &n
	}
	s.created = &out
	return s.created, nil
}
func (s *stub) ListSuggestions(context.Context, string) ([]models.Suggestion, error) {
	return s.suggestions, nil
}
func (s *stub) SetSuggestionStatus(_ context.Context, id, status string) (*models.Suggestion, error) {
	s.statusID, s.statusVal = id, status
	if s.statusErr != nil {
		return nil, s.statusErr
	}
	if s.statusOut != nil {
		return s.statusOut, nil
	}
	return &models.Suggestion{ID: id, Status: status}, nil
}
func (s *stub) ApplySuggestion(_ context.Context, id string) (*store.ApplyResult, error) {
	s.applyID = id
	if s.applyErr != nil {
		return nil, s.applyErr
	}
	return s.applyOut, nil
}

func TestAdminRequiresToken(t *testing.T) {
	h := New(&stub{}, "secret")
	req := httptest.NewRequest(http.MethodGet, "/admin/suggestions", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/suggestions", nil)
	req.Header.Set("X-Admin-Token", "nope")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token code %d", rec.Code)
	}

	h = New(&stub{suggestions: []models.Suggestion{}}, "")
	req = httptest.NewRequest(http.MethodGet, "/admin/suggestions", nil)
	req.Header.Set("X-Admin-Token", "")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("empty server token should lock admin, got %d", rec.Code)
	}
}

func TestAdminApproveRejectApply(t *testing.T) {
	st := &stub{
		statusOut: &models.Suggestion{ID: "abc", Status: "approved", Kind: "new_place"},
		applyOut: &store.ApplyResult{
			Suggestion: &models.Suggestion{ID: "abc", Status: "approved"},
			Place:      &models.Place{ID: "p1", Slug: "m1-test", Name: "M1 Test"},
		},
	}
	h := New(st, "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/suggestions/abc/approve", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || st.statusVal != "approved" || st.statusID != "abc" {
		t.Fatalf("approve %d %s %s", rec.Code, st.statusVal, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/admin/suggestions/abc/reject", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || st.statusVal != "rejected" {
		t.Fatalf("reject %d %s", rec.Code, st.statusVal)
	}

	req = httptest.NewRequest(http.MethodPost, "/admin/suggestions/abc/apply", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || st.applyID != "abc" {
		t.Fatalf("apply %d %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["alreadyApplied"] != false {
		t.Fatalf("body %#v", body)
	}

	req = httptest.NewRequest(http.MethodPatch, "/admin/suggestions/abc", strings.NewReader(`{"status":"rejected"}`))
	req.Header.Set("X-Admin-Token", "secret")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || st.statusVal != "rejected" {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
}

func TestApplyRejectedMapped(t *testing.T) {
	st := &stub{applyErr: &store.Error{Code: 409, Message: "rejected suggestions cannot be applied"}}
	h := New(st, "secret")
	req := httptest.NewRequest(http.MethodPost, "/admin/suggestions/zzz/apply", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 409 {
		t.Fatalf("code %d body %s", rec.Code, rec.Body.String())
	}
}

func TestApplyTwiceIdempotentMapped(t *testing.T) {
	st := &stub{applyOut: &store.ApplyResult{
		AlreadyApplied: true,
		Suggestion:     &models.Suggestion{ID: "abc", Status: "approved"},
		Place:          &models.Place{ID: "p1", Slug: "m1-test", Name: "M1 Test"},
	}}
	h := New(st, "secret")
	req := httptest.NewRequest(http.MethodPost, "/admin/suggestions/abc/apply", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"alreadyApplied":true`) {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestPublicCannotWritePlaces(t *testing.T) {
	h := New(&stub{}, "secret")
	req := httptest.NewRequest(http.MethodPost, "/places", bytes.NewReader([]byte(`{"name":"nope"}`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /places should be 405, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/places/abc", bytes.NewReader([]byte(`{}`)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT /places/{id} should be 405, got %d", rec.Code)
	}
}

func TestContributeDoesNotNeedAdmin(t *testing.T) {
	st := &stub{}
	h := New(st, "secret")
	req := httptest.NewRequest(http.MethodPost, "/suggestions", strings.NewReader(`{
		"kind":"new_place","name":"Warung Tes","notes":"Melawai stall"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
}

func TestCORSAllowsAdminHeaderAndPatch(t *testing.T) {
	h := New(&stub{}, "secret")
	req := httptest.NewRequest(http.MethodOptions, "/admin/suggestions/abc", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	req.Header.Set("Access-Control-Request-Headers", "X-Admin-Token, Content-Type")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("preflight %d", rec.Code)
	}
	allow := rec.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(strings.ToLower(allow), "x-admin-token") {
		t.Fatalf("allow headers %q", allow)
	}
	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "PATCH") {
		t.Fatalf("allow methods %q", methods)
	}
}
