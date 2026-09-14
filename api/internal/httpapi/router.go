package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hansandika/blok-m-msme/api/internal/geofence"
	"github.com/hansandika/blok-m-msme/api/internal/models"
	"github.com/hansandika/blok-m-msme/api/internal/store"
)

// Backend is the persistence surface the HTTP API needs.
type Backend interface {
	ListPlaces(ctx context.Context, f models.PlaceFilters) ([]models.Place, error)
	GetPlace(ctx context.Context, idOrSlug string) (*models.Place, error)
	CreateSuggestion(ctx context.Context, in models.SuggestionInput) (*models.Suggestion, error)
	ListSuggestions(ctx context.Context, status string) ([]models.Suggestion, error)
	SetSuggestionStatus(ctx context.Context, id, status string) (*models.Suggestion, error)
	ApplySuggestion(ctx context.Context, id string) (*store.ApplyResult, error)
}

type Server struct {
	store      Backend
	adminToken string
}

func New(st Backend, adminToken string) http.Handler {
	s := &Server{store: st, adminToken: adminToken}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Admin-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", s.health)
	r.Get("/places", s.listPlaces)
	r.Get("/places/{id}", s.getPlace)
	r.Post("/suggestions", s.createSuggestion)

	r.Route("/admin", func(r chi.Router) {
		r.Use(s.requireAdmin)
		r.Get("/suggestions", s.listSuggestions)
		r.Post("/suggestions/{id}/approve", s.approveSuggestion)
		r.Post("/suggestions/{id}/reject", s.rejectSuggestion)
		r.Post("/suggestions/{id}/apply", s.applySuggestion)
		r.Patch("/suggestions/{id}", s.patchSuggestion)
	})
	return r
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.adminToken == "" || r.Header.Get("X-Admin-Token") != s.adminToken {
			writeError(w, http.StatusUnauthorized, "admin token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "blokm-api",
	})
}

func (s *Server) listPlaces(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	openLate, err := store.ParseBool(q.Get("openLate"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "openLate must be true or false")
		return
	}
	bbox, err := store.ParseBBox(q.Get("bbox"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filters := models.PlaceFilters{
		Query:    strings.TrimSpace(q.Get("q")),
		Category: strings.ToLower(strings.TrimSpace(q.Get("category"))),
		Vibe:     strings.ToLower(strings.TrimSpace(q.Get("vibe"))),
		Price:    strings.TrimSpace(q.Get("price")),
		OpenLate: openLate,
		BBox:     bbox,
	}

	if near := strings.TrimSpace(q.Get("near")); near != "" {
		parts := strings.Split(near, ",")
		if len(parts) != 2 {
			writeError(w, http.StatusBadRequest, "near must be lat,lng")
			return
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 != nil || err2 != nil {
			writeError(w, http.StatusBadRequest, "near must be lat,lng")
			return
		}
		filters.NearLat = &lat
		filters.NearLng = &lng
		if rad := strings.TrimSpace(q.Get("radius")); rad != "" {
			rm, err := strconv.ParseFloat(rad, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "radius must be meters")
				return
			}
			filters.RadiusM = rm
		}
	}

	places, err := s.store.ListPlaces(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list places")
		return
	}
	writeJSON(w, http.StatusOK, models.PlaceList{
		Places:   places,
		Count:    len(places),
		Geofence: geofence.Default(),
	})
}

func (s *Server) getPlace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	place, err := s.store.GetPlace(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load place")
		return
	}
	if place == nil {
		writeError(w, http.StatusNotFound, "place not found")
		return
	}
	writeJSON(w, http.StatusOK, place)
}

func (s *Server) createSuggestion(w http.ResponseWriter, r *http.Request) {
	var in models.SuggestionInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	in.Kind = strings.TrimSpace(in.Kind)
	in.Notes = strings.TrimSpace(in.Notes)
	in.Name = strings.TrimSpace(in.Name)
	in.Category = strings.ToLower(strings.TrimSpace(in.Category))
	in.PlaceID = strings.TrimSpace(in.PlaceID)

	if in.Kind != "new_place" && in.Kind != "correction" {
		writeError(w, http.StatusBadRequest, "kind must be new_place or correction")
		return
	}
	if in.Notes == "" {
		writeError(w, http.StatusBadRequest, "notes are required")
		return
	}
	if in.Kind == "new_place" && in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required for a new place")
		return
	}
	if in.Kind == "correction" && in.PlaceID == "" {
		writeError(w, http.StatusBadRequest, "placeId is required for a correction")
		return
	}
	if in.Category != "" {
		switch in.Category {
		case "japanese", "street", "cafe", "retail", "other":
		default:
			writeError(w, http.StatusBadRequest, "invalid category")
			return
		}
	}

	out, err := s.store.CreateSuggestion(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save suggestion")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"suggestion": out,
		"message":    "Queued for moderation. Live listings are not updated until review.",
	})
}

func (s *Server) listSuggestions(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListSuggestions(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list suggestions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestions": items, "count": len(items)})
}

func (s *Server) approveSuggestion(w http.ResponseWriter, r *http.Request) {
	s.setStatus(w, r, "approved")
}

func (s *Server) rejectSuggestion(w http.ResponseWriter, r *http.Request) {
	s.setStatus(w, r, "rejected")
}

func (s *Server) patchSuggestion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	status := strings.ToLower(strings.TrimSpace(body.Status))
	if status != "approved" && status != "rejected" {
		writeError(w, http.StatusBadRequest, "status must be approved or rejected")
		return
	}
	s.setStatusValue(w, r, status)
}

func (s *Server) setStatus(w http.ResponseWriter, r *http.Request, status string) {
	s.setStatusValue(w, r, status)
}

func (s *Server) setStatusValue(w http.ResponseWriter, r *http.Request, status string) {
	id := chi.URLParam(r, "id")
	out, err := s.store.SetSuggestionStatus(r.Context(), id, status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestion": out})
}

func (s *Server) applySuggestion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := s.store.ApplySuggestion(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	msg := "Applied onto live listings."
	if res.AlreadyApplied {
		msg = "Already applied; live listing unchanged."
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"suggestion":     res.Suggestion,
		"place":          res.Place,
		"alreadyApplied": res.AlreadyApplied,
		"message":        msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeStoreError(w http.ResponseWriter, err error) {
	var se *store.Error
	if errors.As(err, &se) {
		code := se.Code
		if code < 400 || code > 599 {
			code = http.StatusInternalServerError
		}
		writeError(w, code, se.Message)
		return
	}
	writeError(w, http.StatusInternalServerError, "internal error")
}
