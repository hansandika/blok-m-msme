.PHONY: db wait-db api seed web test demo

DATABASE_URL ?= postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable
ADMIN_TOKEN ?= blokm-demo

db:
	docker compose up -d db

wait-db:
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		$(MAKE) db; \
		echo "Waiting for Postgres (docker)..."; \
		n=0; \
		until docker compose exec -T db pg_isready -U blokm -d blokm >/dev/null 2>&1; do \
			n=$$((n+1)); \
			if [ $$n -ge 45 ]; then echo "error: Postgres did not become ready"; exit 1; fi; \
			sleep 1; \
		done; \
		echo "Postgres is ready."; \
	elif pg_isready -h localhost -p 5432 >/dev/null 2>&1; then \
		echo "Using existing Postgres on localhost:5432 (docker not available)."; \
	else \
		echo "error: start Postgres with 'docker compose up -d db', or have a local server on :5432"; \
		exit 1; \
	fi

api:
	cd api && DATABASE_URL=$(DATABASE_URL) AUTO_SEED=1 ADMIN_TOKEN=$(ADMIN_TOKEN) go run ./cmd/server

seed: wait-db
	cd api && DATABASE_URL=$(DATABASE_URL) go run ./cmd/seed

web:
	cd web && npm run dev

test:
	cd api && go test ./...

# One-shot local demo: Postgres up + healthy, migrations + seed.
# Does not start api/web (those are long-running). Print next steps instead.
demo: wait-db
	cd api && DATABASE_URL=$(DATABASE_URL) go run ./cmd/seed
	@if [ ! -f web/.env.local ]; then cp web/.env.example web/.env.local; echo "Wrote web/.env.local from example."; fi
	@if [ ! -d web/node_modules ]; then echo "Installing web dependencies..."; cd web && npm install; fi
	@echo
	@echo "============================================================"
	@echo "Blok M Lokal demo is ready (no paid API keys required)."
	@echo "============================================================"
	@echo
	@echo "Database: $(DATABASE_URL)"
	@echo "Admin token: $(ADMIN_TOKEN)"
	@echo
	@echo "Next steps — two terminals:"
	@echo "  1. make api     # Go API on :8080  (ADMIN_TOKEN=$(ADMIN_TOKEN))"
	@echo "  2. make web     # Next.js on :3000"
	@echo
	@echo "Then open:"
	@echo "  http://localhost:3000           map + filters"
	@echo "  http://localhost:3000/contribute  queue a suggestion (does not write places)"
	@echo "  http://localhost:3000/admin       paste token $(ADMIN_TOKEN)  → approve / apply"
	@echo
	@echo "curl admin queue:"
	@echo "  curl -s -H 'X-Admin-Token: $(ADMIN_TOKEN)' http://localhost:8080/admin/suggestions"
	@echo
	@echo "Seed notes: docs/SEED_AUDIT.md"
