# Blok M Lokal

Opinionated MSME discovery for **Blok M / Melawai / M Bloc / Blok M Hub** in South Jakarta. This is module 1 of a future **Blok M Tonight** product (later: event radar + crowd planner).

See also: [PRD](docs/PRD.md) · [Seed audit](docs/SEED_AUDIT.md)

Locals already drown in ~500 MRT Hub tenants plus Melawai Japanese plus kaki lima. They want filters like *quiet izakaya*, *quick kopi*, *open past 11* — not another Google clone.

The demo runs **without paid API keys**. Listings are a hand-curated seed with lat/lng and tags. Mapbox is optional; Leaflet + OpenStreetMap tiles is the default.

## What this app does (M1)

- Map + list inside a Blok M geofence (Kebayoran Baru / Melawai / Panglima Polim / M Bloc / Square / Hub, plus Cipaku / Senopati / Barito edges)
- Local-first tags: vibe, price (`$` / `$$` / `$$$`), open-late, category (Japanese / street / cafe / retail / other)
- Search + filters (English labels; Indonesian aliases in the seed — try `kopi`, `izakaya`, `seblak`, `larut`, `sepi`)
- Place detail: tags, approximate pin, hours if known, source note
- Favorites in `localStorage` (no auth)
- Contribute: suggest a place or a correction → **moderation queue**. Nothing writes to live listings until an admin **applies** an approved row
- Admin: approve / reject / apply (`ADMIN_TOKEN` + `X-Admin-Token`) — tiny `/admin` page or curl
- Seed of 56 realistic Blok M-area places + `go run ./cmd/seed` / `make demo`

**Skipped on purpose:** bookings, payments, live Google Places sync, accounts, Event Radar, Crowd, Instagram scraping.

## Repo layout

```
/web                 Next.js App Router + TypeScript + Tailwind
/api                 Go HTTP API (chi + pgx)
/docker-compose.yml  Postgres (PostGIS image; queries use lat/lng + bbox)
```

## One-shot demo

No paid keys. From a clean checkout:

```bash
make demo
```

That starts Postgres, waits until it is healthy, runs migrations + seed, writes `web/.env.local` if missing, and prints next steps. Then in two terminals:

```bash
make api    # :8080  ADMIN_TOKEN=blokm-demo (override with ADMIN_TOKEN=...)
make web    # :3000
```

Open [http://localhost:3000](http://localhost:3000). Admin queue: [http://localhost:3000/admin](http://localhost:3000/admin) — paste `blokm-demo`.

Or the long form: `make db`, `make api`, `make seed`, `make web`.

`make demo` stays **docker-first** (or an already-running Postgres on `localhost:5432`). Hosted Supabase is optional — see below.

## Optional: Supabase (hosted Postgres)

Local docker-compose remains the zero-config default. Hosted project **blok-m-msme** is already live:

| | |
|---|---|
| Project | `blok-m-msme` |
| Ref | `sfeebnwtaxnztglvutvd` |
| Region | `ap-southeast-1` |
| URL | https://sfeebnwtaxnztglvutvd.supabase.co |

Schema (`places` + `suggestions`, including `applied_at` / `applied_place_id`) is applied, **RLS is enabled**, and **56 seed places** are loaded. Copy the database password from the Supabase dashboard (Settings → Database). **Do not commit it.**

**Use the session pooler** (port `5432`) with Go/pgx. Session mode keeps prepared statements working. Transaction-mode pooler (port `6543`) can break `pgx` prepared statements — skip it for this API.

```bash
# Session pooler (preferred). Direct db.*.supabase.co is IPv6-only and fails on many networks.
export DATABASE_URL='postgresql://postgres.sfeebnwtaxnztglvutvd:[PASSWORD]@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require'
export AUTO_SEED=0   # schema + 56 seed places are already loaded on this project
export ADMIN_TOKEN=blokm-demo
cd api && go run ./cmd/server
```

- Same SQL as local: migrations are `CREATE … IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS`. PostGIS is **not** required (queries use `lat`/`lng`).
- The Next.js app talks only to the Go API (`NEXT_PUBLIC_API_URL`). Do not put `DATABASE_URL` in the browser.
- Pointing `make api` at Supabase: `DATABASE_URL='…' AUTO_SEED=0 make api`
- Fresh empty Supabase DB: `DATABASE_URL='…' make seed` (skips local docker wait) or `cd api && go run ./cmd/seed`. This project is already migrated + seeded, so keep `AUTO_SEED=0`.
- RLS is enabled on the hosted tables. The Go API should use the `postgres` role (or another role that bypasses RLS). Client-side Supabase keys are not used.

`make demo` does **not** target Supabase; it only prepares a local database.

## How to run (manual)

### 1. Postgres

```bash
docker compose up -d
```

Wait until healthy (`docker compose ps`). Default:

`postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable`

The PostGIS image is for later geo work. Bounding box / haversine on `lat`/`lng` is enough.

### 2. API

```bash
cd api
export DATABASE_URL=postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable
export AUTO_SEED=1
export ADMIN_TOKEN=blokm-demo
go run ./cmd/server
```

Listens on `:8080`. If the `places` table is empty, the server auto-loads the seed. To refresh listings after editing JSON:

```bash
cd api && go run ./cmd/seed
```

Seed file: [`api/internal/fsdata/data/places.json`](api/internal/fsdata/data/places.json) — see [seed audit](docs/SEED_AUDIT.md) for counts, gaps, and the “hours/pins are approximate” disclaimer.

### 3. Web

```bash
cd web
cp .env.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Environment

| Variable | Where | Default | Notes |
|---|---|---|---|
| `DATABASE_URL` | API | `postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable` | Local docker default. For hosted Supabase use the **session pooler** URI (`sslmode=require`). Never commit a password. |
| `API_ADDR` | API | `:8080` | |
| `AUTO_SEED` | API | `1` | Set `0` to skip seeding (use `0` on the already-seeded Supabase project) |
| `ADMIN_TOKEN` | API | unset in the binary; `blokm-demo` via `make api` / `make demo` | Required for `/admin/*`. Header: `X-Admin-Token` |
| `NEXT_PUBLIC_API_URL` | web | `http://localhost:8080` | |
| `NEXT_PUBLIC_MAPBOX_TOKEN` | web | unset | If set, Mapbox GL JS dark map. Otherwise Leaflet + OSM (no key) |

Copy [`.env.example`](.env.example). Do not commit real tokens.

## API

- `GET /health`
- `GET /places?q=&category=&vibe=&price=&openLate=&bbox=west,south,east,north`
- `GET /places?near=lat,lng&radius=1200` (meters, default 1200)
- `GET /places/:id` — UUID or slug
- `POST /suggestions` — `{ kind: "new_place"|"correction", name?, placeId?, category?, notes, lat?, lng? }` (public; **does not** write `places`)
- `GET /admin/suggestions?status=pending|approved|rejected|applied` — `X-Admin-Token`
- `POST /admin/suggestions/:id/approve`
- `POST /admin/suggestions/:id/reject`
- `POST /admin/suggestions/:id/apply` — copy an **approved** row onto live `places`
- `PATCH /admin/suggestions/:id` — `{ "status": "approved"|"rejected" }`

There is no public `POST`/`PUT`/`PATCH` on `/places`.

Categories: `japanese` `street` `cafe` `retail` `other`  
Vibes: `quiet` `lively` `hidden` `hangout` `date-night` `after-office` `quick-bite` `family`  
Price: `$` `$$` `$$$`

## Admin apply flow

1. Someone submits `/contribute` (or `POST /suggestions`). Row lands in `suggestions` as `pending`.
2. Moderator lists the queue (`GET /admin/suggestions` or `/admin` in the web app).
3. **Approve** or **reject** (status only). Rejected rows never change `places`.
4. **Apply** an approved row:
   - `new_place` → insert a place (fresh UUID; slug from the name, uniqued on collision; geofence checked when a pin is present; missing fields get defaults: category `other`, price `$$`, pin = geofence center, neighborhood `Blok M`)
   - `correction` → update the referenced place from structured fields plus notes (see below)
5. Applying twice returns the same place with `alreadyApplied: true` (idempotent). Applied rows cannot change status.

### Notes overlay (corrections and extra new-place fields)

Key:value lines in `notes` (case-insensitive keys), one per line:

`name`, `category`, `price`, `hours`, `neighborhood`, `address`, `description`, `source` / `sourceNote`, `vibe` / `vibes` (comma list of known tags), `openLate` / `open-late` / `larut` (`true`/`false`), `lat`, `lng`

Bare tokens if no key was set: `open late` / `larut` → `openLate`; `not open late` / `closes early` → not late; a lone `$` / `$$` / `$$$` → price.

If a correction has no structured fields, the notes are appended to the place description as a community comment so the report is not discarded.

```bash
# list
curl -s -H "X-Admin-Token: blokm-demo" http://localhost:8080/admin/suggestions

# approve, then apply
curl -s -X POST -H "X-Admin-Token: blokm-demo" http://localhost:8080/admin/suggestions/SUGGESTION_ID/approve
curl -s -X POST -H "X-Admin-Token: blokm-demo" http://localhost:8080/admin/suggestions/SUGGESTION_ID/apply
```

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
- **Blok M Tonight:** Event Radar + Crowd on the same geofence and place IDs.

## License / data

Seed pins and hours are **approximate**, hand-curated for a demo, not scraped from Instagram and not a live Google dump. Confirm on-site before you go.
