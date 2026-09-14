package store

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Kopi Tuku Cipaku":     "kopi-tuku-cipaku",
		"  Ramen 38Sanpachi  ": "ramen-38sanpachi",
		"$$$":                  "place",
		"Warung Nasi Hub":      "warung-nasi-hub",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q)=%q want %q", in, got, want)
		}
	}
}

func TestParseNotesKeyValues(t *testing.T) {
	p := ParseNotes(`hours: 17:00–01:00
open-late: true
price: $$
neighborhood: Melawai
vibes: quiet, date-night
category: japanese
lat: -6.2442
lng: 106.7995`)
	if p.Hours == nil || *p.Hours != "17:00–01:00" {
		t.Fatalf("hours: %#v", p.Hours)
	}
	if p.OpenLate == nil || !*p.OpenLate {
		t.Fatal("openLate")
	}
	if p.Price == nil || *p.Price != "$$" {
		t.Fatalf("price %#v", p.Price)
	}
	if p.Neighborhood == nil || *p.Neighborhood != "Melawai" {
		t.Fatal("neighborhood")
	}
	if p.Category == nil || *p.Category != "japanese" {
		t.Fatal("category")
	}
	if len(p.Vibes) != 2 {
		t.Fatalf("vibes %#v", p.Vibes)
	}
	if p.Lat == nil || p.Lng == nil {
		t.Fatal("coords")
	}
}

func TestParseNotesBareTokens(t *testing.T) {
	p := ParseNotes("Still open late / larut on Melawai. Price $$")
	if p.OpenLate == nil || !*p.OpenLate {
		t.Fatal("expected open late")
	}
	if p.Price == nil || *p.Price != "$$" {
		t.Fatalf("price %#v", p.Price)
	}

	p = ParseNotes("Not open late anymore, closes early.")
	if p.OpenLate == nil || *p.OpenLate {
		t.Fatal("expected not open late")
	}
}

func TestParseNotesKeyBeatsBareToken(t *testing.T) {
	p := ParseNotes("openLate: false\nThey used to be larut but not now.")
	if p.OpenLate == nil || *p.OpenLate {
		t.Fatal("key:value should win over bare larut")
	}
}

func TestResolveCoordsGeofence(t *testing.T) {
	outsideLat, outsideLng := -6.1, 106.8
	_, _, err := resolveCoords(&outsideLat, &outsideLng, nil, nil, true)
	if err == nil {
		t.Fatal("expected geofence error")
	}
	lat, lng, err := resolveCoords(nil, nil, nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if lat == 0 || lng == 0 {
		t.Fatalf("expected center defaults, got %v %v", lat, lng)
	}
	onlyLat := -6.244
	_, _, err = resolveCoords(&onlyLat, nil, nil, nil, true)
	if err == nil {
		t.Fatal("expected lat/lng pair error")
	}
}
