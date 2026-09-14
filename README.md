# Blok M Lokal

Opinionated MSME discovery for **Blok M / Melawai / M Bloc / Blok M Hub** in South Jakarta. This is module 1 of a future **Blok M Tonight** product (later: event radar + crowd planner).

See also: [PRD](docs/PRD.md)

Locals already drown in ~500 MRT Hub tenants plus Melawai Japanese plus kaki lima. They want filters like *quiet izakaya*, *quick kopi*, *open past 11* — not another Google clone.

The demo runs **without paid API keys**. Listings are a hand-curated seed with lat/lng and tags. Mapbox is optional; Leaflet + OpenStreetMap tiles is the default.

## What this MVP does

- Map + list inside a Blok M geofence (Kebayoran Baru / Melawai / Panglima Polim / M Bloc / Square / Hub, plus Cipaku / Senopati / Barito edges)
- Local-first tags: vibe, price (`$` / `$$` / `$$$`), open-late, category (Japanese / street / cafe / retail / other)
- Search + filters (English labels; Indonesian aliases in the seed — try `kopi`, `izakaya`, `seblak`, `larut`, `sepi`)
- Place detail: tags, approximate pin, hours if known, source note
- Favorites in `localStorage` (no auth)
- Contribute: suggest a place or a correction → **moderation queue table**. Nothing writes to live listings without review
- Seed of 56 realistic Blok M-area places + `go run ./cmd/seed`

**Skipped on purpose:** bookings, payments, live Google Places sync, accounts, crowd/events.

## Repo layout

```
/web                 Next.js App Router + TypeScript + Tailwind
/api                 Go HTTP API (chi + pgx)
/docker-compose.yml  Postgres (PostGIS image; MVP queries use lat/lng + bbox)
```

## How to run

### 1. Postgres

```bash
docker compose up -d
```

Wait until healthy (`docker compose ps`). Default:

`postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable`

The PostGIS image is for later geo work. The MVP does **not** require PostGIS functions — bounding box / haversine on `lat`/`lng` is enough.

### 2. API

```bash
cd api
export DATABASE_URL=postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable
export AUTO_SEED=1
go run ./cmd/server
```

Listens on `:8080`. If the `places` table is empty, the server auto-loads the seed. To refresh listings after editing JSON:

```bash
cd api && go run ./cmd/seed
```

Seed file: [`api/internal/fsdata/data/places.json`](api/internal/fsdata/data/places.json)

### 3. Web

```bash
cd web
cp .env.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

Or from the repo root: `make db`, `make api`, `make seed`, `make web`.

## Environment

| Variable | Where | Default | Notes |
|---|---|---|---|
| `DATABASE_URL` | API | `postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable` | |
| `API_ADDR` | API | `:8080` | |
| `AUTO_SEED` | API | `1` | Set `0` to skip seeding an empty DB |
| `ADMIN_TOKEN` | API | unset | If set, `GET /admin/suggestions` requires `X-Admin-Token` |
| `NEXT_PUBLIC_API_URL` | web | `http://localhost:8080` | |
| `NEXT_PUBLIC_MAPBOX_TOKEN` | web | unset | If set, Mapbox GL JS dark map. Otherwise Leaflet + OpenStreetMap tiles (no key; tiles are inverted to match the dark UI) |

Copy [`.env.example`](.env.example). Do not commit real tokens.

## API

- `GET /health`
- `GET /places?q=&category=&vibe=&price=&openLate=&bbox=west,south,east,north`
- `GET /places?near=lat,lng&radius=1200` (meters, default 1200)
- `GET /places/:id` — UUID or slug
- `POST /suggestions` — `{ kind: "new_place"|"correction", name?, placeId?, category?, notes, lat?, lng? }`
- `GET /admin/suggestions?status=pending` — only with `ADMIN_TOKEN`

Categories: `japanese` `street` `cafe` `retail` `other`  
Vibes: `quiet` `lively` `hidden` `hangout` `date-night` `after-office` `quick-bite` `family`  
Price: `$` `$$` `$$$`

## Bilingual search notes

UI chrome is English. Seed `aliases` and copy include Indonesian so these queries work:

| Try | Tends to surface |
|---|---|
| `kopi` / `quick kopi` | cafes, Tuku, Kenangan |
| `izakaya` / `sepi` / `quiet` | quieter Japanese rooms |
| `larut` / `open late` | Melawai night + open-late flag |
| `seblak` / `kaki lima` / `martabak` | street |
| `m bloc` / `hub` | compound / MRT tenants |

## Future hooks (not built)

- **Google Places:** keep `source_note` + stable `slug`. A later job can match on name+geofence and refresh hours without replacing local tags.
- **Community edits:** `suggestions` is the queue (`pending` / `approved` / `rejected`). An admin tool should apply approved rows onto `places` — never public `INSERT` into live listings.
- **Blok M Tonight:** events + crowd on the same geofence and place IDs.

## License / data

Seed pins and hours are **approximate**, hand-curated for a demo, not scraped from Instagram and not a live Google dump. Confirm on-site before you go.
