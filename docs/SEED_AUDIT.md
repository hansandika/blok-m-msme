# Seed audit — Blok M Lokal v1.1

Hand-curated `places.json` for the local demo. **Hours and pins are approximate.** This is not a live Google dump and not scraped from Instagram. Confirm on-site before you go.

Source file: [`api/internal/fsdata/data/places.json`](../api/internal/fsdata/data/places.json)

Geofence (product law): south `-6.2560`, north `-6.2340`, west `106.7900`, east `106.8105` (Melawai / Panglima Polim / M Bloc / Square / Hub + Cipaku / Senopati / Barito edges).

## Counts (56 places)

### Category

| Category | Count |
|---|---|
| japanese | 18 |
| street | 14 |
| cafe | 12 |
| retail | 6 |
| other | 6 |

### Neighborhood

| Neighborhood | Count |
|---|---|
| Melawai | 18 |
| Blok M Plaza | 11 |
| M Bloc | 7 |
| Blok M Hub | 6 |
| Blok M Square | 5 |
| Panglima Polim | 4 |
| Senopati | 2 |
| Cipaku | 1 |
| Barito | 1 |
| Pasaraya | 1 |

### Open late / larut

| Open late | Count |
|---|---|
| false | 44 |
| true | 12 |

Open-late is the night path: Melawai izakaya + kaki lima after 23:00, plus M Bloc garden. Mall / Hub tenants are mostly daytime–evening. A v1.1 pass unmarked **Ikkudo Ichi** (`11:00–22:30`) which had been flagged open-late despite closing before 23:00.

### Price

| Price | Count |
|---|---|
| `$` | 29 |
| `$$` | 24 |
| `$$$` | 3 |

## What is approximate

- **Pins** sit on the right street / compound, not a surveyed doorway. Several Hub and Plaza rows share a compound centroid on purpose.
- **Hours** are demo-grade (typical trade hours), not scraped and not guaranteed tonight.
- **Composite rows** stand in for density the seed cannot enumerate: Hub Tenant Snack Kiosk, Blok M Hub Food Hall, Local Fashion Kiosk M Bloc, Distro & Vinyl M Bloc, M Bloc Market.

## Known gaps

- Blok M Hub has on the order of **~500 tenants**; the seed has **6** Hub rows. Discovery overload is the product problem — we did not try to list them all.
- **Barito, Cipaku, Pasaraya, Senopati** are thin (1–2 each). Edge coverage is a sketch, not a directory.
- **Retail** is thin (6). Distro / vinyl / market are compound-level, not shop-level.
- **Quiet izakaya** and **open-late street** are stronger than family-mall Japanese.
- No events, crowd signals, or live hours — those are later modules (Event Radar / Crowd).
- Re-seeding **upserts by slug** and will not delete community places created via admin apply.

## How to re-seed

Empty DB is auto-seeded when the API starts with `AUTO_SEED=1` (default).

To reload JSON onto an existing DB (upsert by `slug`):

```bash
make seed
# or
cd api && DATABASE_URL=postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable go run ./cmd/seed
```

`make demo` brings Postgres up, waits until healthy, then runs this seed.

To point the API at the hosted Supabase project instead, set `DATABASE_URL` to the **session pooler** URI and `AUTO_SEED=0` (that database is already seeded). See the README. Never commit the password.

Do **not** public-write `places`. New listings enter through `POST /suggestions` → admin approve → admin apply.
