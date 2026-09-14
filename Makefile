.PHONY: db api seed web test

db:
	docker compose up -d db

api:
	cd api && DATABASE_URL=$${DATABASE_URL:-postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable} AUTO_SEED=1 go run ./cmd/server

seed:
	cd api && DATABASE_URL=$${DATABASE_URL:-postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable} go run ./cmd/seed

web:
	cd web && npm run dev

test:
	cd api && go test ./...
