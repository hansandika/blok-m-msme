package geofence

// Approximate Blok M / Kebayoran Baru working box:
// Melawai, Panglima Polim, M Bloc, Blok M Square, Hub, Barito, Cipaku, Senopati south edge.
const (
	South     = -6.2560
	North     = -6.2340
	West      = 106.7900
	East      = 106.8105
	CenterLat = -6.2442
	CenterLng = 106.7995
)

type Box struct {
	South float64 `json:"south"`
	North float64 `json:"north"`
	West  float64 `json:"west"`
	East  float64 `json:"east"`
}

func Default() Box {
	return Box{South: South, North: North, West: West, East: East}
}

func Contains(lat, lng float64) bool {
	return lat >= South && lat <= North && lng >= West && lng <= East
}
