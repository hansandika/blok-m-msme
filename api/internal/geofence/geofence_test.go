package geofence

import "testing"

func TestContains(t *testing.T) {
	if !Contains(CenterLat, CenterLng) {
		t.Fatal("center should be inside")
	}
	if Contains(-6.10, 106.80) {
		t.Fatal("Menteng-ish pin should be outside")
	}
	if Contains(CenterLat, 106.70) {
		t.Fatal("west of geofence")
	}
}
