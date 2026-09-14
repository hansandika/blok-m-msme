package models

import "time"

type Place struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Vibes        []string `json:"vibes"`
	Price        string   `json:"price"`
	OpenLate     bool     `json:"openLate"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Neighborhood string   `json:"neighborhood"`
	Address      string   `json:"address"`
	Hours        *string  `json:"hours,omitempty"`
	Description  string   `json:"description"`
	SourceNote   string   `json:"sourceNote"`
	Aliases      []string `json:"aliases"`
}

type PlaceSeed struct {
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Vibes        []string `json:"vibes"`
	Price        string   `json:"price"`
	OpenLate     bool     `json:"openLate"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Neighborhood string   `json:"neighborhood"`
	Address      string   `json:"address"`
	Hours        string   `json:"hours"`
	Description  string   `json:"description"`
	SourceNote   string   `json:"sourceNote"`
	Aliases      []string `json:"aliases"`
}

type Suggestion struct {
	ID             string     `json:"id"`
	Kind           string     `json:"kind"`
	PlaceID        *string    `json:"placeId,omitempty"`
	Name           *string    `json:"name,omitempty"`
	Category       *string    `json:"category,omitempty"`
	Notes          string     `json:"notes"`
	Lat            *float64   `json:"lat,omitempty"`
	Lng            *float64   `json:"lng,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	AppliedAt      *time.Time `json:"appliedAt,omitempty"`
	AppliedPlaceID *string    `json:"appliedPlaceId,omitempty"`
}

type SuggestionInput struct {
	Kind     string   `json:"kind"`
	PlaceID  string   `json:"placeId"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Notes    string   `json:"notes"`
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
}

type PlaceList struct {
	Places   []Place `json:"places"`
	Count    int     `json:"count"`
	Geofence any     `json:"geofence"`
}

type PlaceFilters struct {
	Query    string
	Category string
	Vibe     string
	Price    string
	OpenLate *bool
	BBox     *BBox
	NearLat  *float64
	NearLng  *float64
	RadiusM  float64
}

type BBox struct {
	West  float64
	South float64
	East  float64
	North float64
}
