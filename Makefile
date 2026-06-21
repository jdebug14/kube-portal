# ── Variables ────────────────────────────────────────────────
BINARY_NAME=kube-portal
IMAGE_NAME=kube-portal:dev

# ── Dev ───────────────────────────────────────────────────────
.PHONY: dev start-server start-web

start-dev:
	@cd web && npx concurrently \
		--names "api,web" \
		--prefix-colors "magenta,green" \
		"cd ../api && go run ./cmd/server" \
		"npm run dev"

start-server:
	@echo "→ Starting API server..."
	@cd api && go run ./cmd/server

start-web:
	@echo "→ Starting web server..."
	@cd web && npm run dev