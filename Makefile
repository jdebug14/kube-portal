# ── Variables ────────────────────────────────────────────────
BINARY_NAME=kube-portal
IMAGE_NAME=kube-portal:dev

# ── Dev environment ───────────────────────────────────────────

.PHONY: env-start
env-start:
	@echo "→ Opening VS Code"
	@code .
	@echo "→ Starting Docker..."
	@sudo service docker start 2>/dev/null || true
	@echo "→ Waiting for Docker to be ready..."
	@until docker info >/dev/null 2>&1; do sleep 1; done
	@echo "→ Starting Minikube..."
	@minikube start --driver=docker 2>/dev/null || true
	@echo "✓ Environment ready"

.PHONY: env-stop
env-stop:
	@echo "→ Stopping Minikube..."
	@minikube stop
	@echo "→ Stopping Docker..."
	@sudo service docker stop
	@echo "✓ Environment stopped"

.PHONY: dev
dev: env-start
	@echo "→ Starting API and frontend..."
	# We'll fill this in when we have actual code to run

.PHONY: deploy-local
deploy-local: env-start
	@echo "→ Building image..."
	# We'll fill this in when we have a Dockerfile

.PHONY: status
status:
	@echo "=== Docker ===" && docker info --format '{{.ServerVersion}}' 2>/dev/null && echo "running" || echo "stopped"
	@echo "=== Minikube ===" && minikube status 2>/dev/null || echo "stopped"