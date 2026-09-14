# Blok M Tonight — PRD

**One-liner:** Help people decide *where to go around Blok M tonight* — places first, then events and crowd — without another Google clone or paid APIs for the local demo.

**Product shape:** one app, three modules on the same geofence + place IDs.

1. **Blok M Lokal** (MSME discovery) — module 1 / current MVP
2. **Event Radar** — what’s on tonight / this weekend nearby
3. **Crowd / go-time** — when a spot is worth going (or skip)

---

## Problem

Blok M is dense (~500 Hub tenants + Melawai Japanese + kaki lima + M Bloc). Locals don’t need “restaurants near me.” They need *quiet izakaya*, *quick kopi*, *open past 11*, *not dead at 9*. Later they also need “what’s happening” and “is it packed yet.”

## Goals

- Opinionated discovery inside a hard Blok M geofence (Melawai / Panglima Polim / M Bloc / Square / Hub + edges)
- Tags that match how people talk (vibe, price, open-late, category + bilingual aliases)
- Local demo path with **zero required paid keys** (Leaflet/OSM default; Mapbox optional)
- Curated seed + human moderation queue — never scrape Instagram, never public-write live listings
- Shared place identity so Event Radar and Crowd plug into the same map later

## Non-goals (v1–v2)

Bookings, payments, accounts (beyond optional later), live Google Places as source of truth, social feeds, city-wide coverage, Instagram scraping.

## Primary user

South Jakarta locals / regulars planning tonight or a quick after-office stop. Secondary: visitors who want a shortlist, not a directory dump.

## Success metrics (demo-grade)

- Cold start → useful shortlist in under 30s (filter or search)
- Seed coverage feels real for Melawai / M Bloc / Hub night paths (≥50 places, mixed categories)
- Suggest-a-place never mutates live data without review
- `docker compose` + API + web runs without external API keys

---

## Module 1 — Blok M Lokal (current MVP)

**Status:** implemented on branch/PR for Blok M Lokal MSME discovery MVP.

| Capability | Spec |
|---|---|
| Map + list | Sync selection; geofence bbox / near+radius |
| Filters | category, vibe, price `$`/`$$`/`$$$`, open-late, `q` (aliases) |
| Detail | tags, pin, hours if known, source note |
| Favorites | `localStorage` only |
| Contribute | `new_place` / `correction` → `suggestions` (`pending`/`approved`/`rejected`) |
| Seed | ~56 hand-curated places; auto-seed empty DB |
| Stack | Next.js App Router + TS + Tailwind; Go/chi + Postgres; PostGIS image optional |

**MVP done when:** smoke test plan is green locally; README run path works; PR merged to `main`.

**v1.1 hardening (post-merge):** admin apply-approved-suggestion flow; seed QA pass (hours/pins); light E2E smoke; optional Mapbox polish.

---

## Module 2 — Event Radar

**Job:** “What’s on around Blok M tonight / this weekend?”

**v0 scope**

- `events` table: title, venue/place_id (nullable), lat/lng, start/end, source, tags
- Hand-curated seed (M Bloc / Hub / Melawai nights) — same no-paid-key rule
- UI: date chips (Tonight / Weekend), list + map pins, link into place detail when `place_id` set
- Ingest hook later (ICS / manual CSV / partner feed) — not Instagram

**Out of scope for v0:** ticketing, RSVP, full city calendars.

---

## Module 3 — Crowd / go-time

**Job:** “Is it worth going *now*?”

**v0 scope (honest, not fake live)**

- Per-place `crowd_signals`: daypart buckets (e.g. Thu–Sat 18–22) with curated `typical` level + optional user check-in chip (“quiet / ok / packed”)
- Surface on place detail + optional map heat tint
- Label clearly as *typical / community*, not live CCTV

**Later:** optional anonymous check-ins with decay; never require accounts for read.

---

## Architecture principles

- One Postgres, shared `places.id` / `slug` across modules
- Geofence is product law — queries always scoped
- Local tags beat vendor categories; Google (if ever) only refreshes hours/address via `source_note` + slug match
- Paid keys optional forever for the happy local path
- Contribute → queue → admin apply; no public `INSERT` on `places`

---

## Milestones

| # | Milestone | Outcome | Exit criteria |
|---|---|---|---|
| **M0** | MVP merge | Blok M Lokal usable locally | MVP PR undrafted, smoke test plan checked, merged to `main` |
| **M1** | Lokal v1.1 | Trustworthy seed + moderation | Admin approve/reject/apply; seed audit notes; Makefile one-shot `make demo` |
| **M2** | Event Radar v0 | Tonight/weekend list on same map | Seed ≥15 events; filter by night; place deep-links work |
| **M3** | Crowd v0 | Go-time cue on detail + map | Daypart typical levels for top ~30 places; community chip writes signal |
| **M4** | Blok M Tonight shell | One nav: Places / Events / Tonight | Shared header, deep links, “Tonight” composite view (open-late + events + quieter picks) |
| **M5** | Soft launch path | Shareable demo | Deploy story (still key-optional); short demo script; contribution loop documented |

**Suggested near-term order:** finish M0 → M1 in parallel with PRD freeze → M2 before M3 (events are clearer content) → M4 stitches product → M5 only when demo is sticky.

---

## Open decisions (defaults)

1. **Brand in UI:** “Blok M Lokal” for module 1; product umbrella “Blok M Tonight” from M4.
2. **Auth:** none until admin tools need it; admin = `ADMIN_TOKEN` header for now.
3. **Events source v0:** curated JSON seed, same pattern as places.
4. **Crowd v0:** curated typical + optional anonymous chip; no accounts.
