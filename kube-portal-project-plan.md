# kube-portal — project plan

## Philosophy

> Ship something a strong Platform Engineer would be proud of. Not perfect, but a solid foundation with clear seams for future iteration.

- **Move fast, don't rush.** Velocity comes from good decisions early, not from skipping them.
- **Reason about tradeoffs, then decide and move.** We'll note why we made each call — not to over-engineer, but to learn.
- **Strong foundations over polish.** Correct error handling, sensible structure, and good defaults matter more than UI shine right now.
- **Observability over operations.** Read before you write. Trust is earned with a portal that shows you the truth before it lets you change things.

---

## Scope (v1)

### In scope
- Namespace selector with cluster-wide view
- Deployments, ReplicaSets, and Pods — list and detail views
- Pod log tail — last N lines, static fetch (no streaming)
- Kubernetes Events feed
- Pod CPU and memory metrics via Metrics Server (graceful degradation if absent)
- Kubernetes manifests: Deployment, Service, ServiceAccount, ClusterRole, ClusterRoleBinding
- Single Docker image (Go binary + embedded React assets)

### Explicitly out of scope (v1)
| Item | Reason | Future milestone |
|---|---|---|
| Helm chart | Packaging, not functionality | v2 |
| OIDC / SSO auth | Mini-project on its own | v2 |
| Restart / scale operations | Mutations need more hardening and RBAC care | v2 |
| Multi-cluster support | Complexity multiplier; nail single-cluster first | v3 |
| Prometheus / custom metrics | No external dependencies in v1 | v2 |
| Dark mode / theming | UI polish | v2 |
| Live log streaming | Solved better by dedicated tooling (Grafana/Loki, Datadog); static tail covers immediate need | v2 if justified |


### Moved out of scope (v1)
#### Pod metrics
- `GET /api/v1/namespaces/:ns/pods/:pod/metrics` — via `metrics.k8s.io/v1beta1`
- Graceful degradation: if Metrics Server absent, API returns `metrics_available: false` and UI shows a callout instead of an error
- CPU (millicores) and memory (MiB) with request/limit context where available
---

## Timeline

### Week 1 — Skeleton + Workloads

**Goal:** A working dev loop and the first real data on screen.

#### Day 1: Repo scaffold
- `git init kube-portal` with `.gitignore`, `README.md`, `Makefile`
- Go module: `api/` with `cmd/server/main.go`, `/healthz` endpoint

#### Day 2–3: Kubernetes client + Workloads API
- `internal/k8s/client.go` — `client-go` auto-detect: tries in-cluster service account token first, falls back to `~/.kube/config` for out-of-cluster operation (local dev). Same binary, zero code changes between environments.
- Handlers: `GET /api/v1/namespaces`, `GET /api/v1/namespaces/:ns/deployments`, `GET /api/v1/namespaces/:ns/pods`
- Typed response structs — no raw k8s objects leaked to the frontend

**Design decision to consider:** _Typed response structs vs. forwarding raw k8s JSON._ Forwarding is faster to write but tightly couples the frontend to the k8s API shape. Typed structs give us a stable contract and let us evolve independently. We'll use typed structs with explicit field mapping.

#### Day 4: Workloads UI
**Design decision to consider:** _How do we serve the frontend in production?_ Options: (a) embed React build into Go binary with `embed.FS`, (b) separate nginx container. We'll go with (a) — single binary is simpler to deploy on any cluster, no sidecar to manage.

- Namespace selector (global, persisted in URL param)
- Deployments list: name, namespace, ready/desired, age, image
- Pods list: name, status, restarts, node, age
- Pod detail page: labels, annotations, conditions, container statuses

**Checkpoint:** Can browse namespaces, see deployments and pods, click into a pod.

---

### Week 2 — More Observability

**Goal:** A portal that gives you real signal — logs and events — the two things you reach for first in any incident.

#### Day 1–2: Log tail
- `GET /api/v1/namespaces/:ns/pods/:pod/logs?tail=200` — returns last N lines as plain text via `client-go`
- Tail line count configurable via query param, capped at a sensible max (e.g. 1000)
- Handles pod-not-found and container-not-ready cleanly — typed errors, not panics
- Frontend: `LogViewer` component — monospace, scrollable, "Copy to clipboard" button, tail count selector
- Link to external log platform (e.g. Grafana) configurable via env var — shown as a callout in the UI

**Decision made:** _Static tail vs. live streaming._ In any mature setup, logs belong in a dedicated system (Grafana/Loki, Datadog, Elastic). Building a full streaming implementation here solves a problem that's already solved better elsewhere. Static tail is genuinely useful for quick checks ("is this pod doing anything?") without the complexity of WebSocket lifecycle, reconnection logic, or distinguishing network blips from pod termination. Live streaming deferred to v2 if there's a clear use case that external tooling doesn't cover.

#### Day 3: Events feed
- `GET /api/v1/namespaces/:ns/events` — filtered by involved object where applicable
- Warning events surfaced prominently (colour-coded in UI)
- Sorted by last timestamp descending
- Pod detail page shows related events inline

#### Day 4-5: Error handling + resilience
- Consistent API error envelope: `{ error: string, code: number }`
- Frontend error boundaries on every data-fetching page
- Loading skeletons (not spinners) — better perceived performance
- Handle common failure modes (e.g. namespace not found, pod already terminated, metrics unavailable)

**Checkpoint:** Can view pod logs and events, degrade gracefully and get meaning error messages.

---

### Week 3 — Hardening + Deployable Artifact

**Goal:** Something you'd actually hand to a team to run. Not pretty, but trustworthy.

#### Day 1-2: Automated Testing
- client-go ships a fake clientset (k8s.io/client-go/kubernetes/fake) that lets you seed test data and call your functions against it without a real cluster.
  - test things like "given a deployment with nil Replicas, DesiredReplicas maps to 0"
- Handler integration tests:
  - Verify the full path (routing → handler → client → JSON encoding) using `httptest.NewRecorder()` and `httptest.NewRequest()` with the fake client wired in.
- Frontend:
    - Vitest (natural fit, shares Vite config), React Testing Library, jsdom as the test environment, and MSW (Mock Service Worker) for API mocking. MSW is worth the setup — it intercepts at the network level rather than mocking fetch directly, which means tests exercise the actual fetch logic in your query functions.
- Github actions to automate test execution on every commit
- Documentation: definintely Go doc comments, maybe API spec

#### Day 3: RBAC + security posture
- RBAC: `ClusterRole` with `get/list/watch` on `namespaces`, `deployments`, `replicasets`, `pods`
- Least-privilege `ClusterRole` — only the verbs we actually use
- Non-root container, read-only root filesystem in the Deployment manifest
- Resource requests and limits on the portal Pod
- `SecurityContext`: `runAsNonRoot: true`, `allowPrivilegeEscalation: false`

**Design decision to consider:** _ClusterRole vs. Roles per namespace._ ClusterRole is simpler but broad. Namespace-scoped Roles are more secure but require enumerating namespaces upfront or a separate ClusterRole just for namespace listing. For v1 we use ClusterRole with the minimum required verbs — note this in the README as a known tradeoff.

#### Day 4: Docker image + manifests
- Multi-stage Dockerfile: `node` build stage → `go build` stage → minimal `distroless` runtime
- Image size target: under 50MB
- Complete `deploy/manifests/` — everything needed to `kubectl apply -f deploy/manifests/`
- Configurable via environment variables: `KUBEPORTAL_PORT`, `KUBEPORTAL_LOG_LEVEL`
  - React scaffold: `web/` with Vite + TypeScript + TanStack Router + React Query
- `make dev` runs both API and frontend with hot reload
- `make deploy-local` builds image, loads into Minikube (`minikube image load`), and applies manifests — primary dev workflow
- `deploy/manifests/` skeleton: `namespace.yaml`, `serviceaccount.yaml`

**Dev workflow note:** Primary development targets running the portal as a Pod inside Minikube — this exercises the real deployment path and RBAC from day one. Out-of-cluster operation (plain `go run`, reading `~/.kube/config`) is supported automatically by `client-go`'s auth auto-detect and is useful as a fallback. `make dev` uses out-of-cluster for fast iteration; `make deploy-local` uses in-cluster for deployment testing.

#### Day 5: README + wrap-up
- Getting started (local dev, in-cluster deploy)
- Architecture overview (one paragraph + the architecture diagram)
- Known limitations and v2 roadmap items
- Final review: does it work on a fresh cluster?

**Checkpoint:** `kubectl apply -f deploy/manifests/` on a real cluster, `kubectl port-forward`, portal is usable.

---

## Tech stack

| Layer | Choice | Rationale |
|---|---|---|
| Backend language | Go | Mature k8s client ecosystem, single binary, fast |
| k8s client | `client-go` | Official, battle-tested, handles in-cluster auth |
| HTTP router | `chi` | Lightweight, idiomatic, good middleware story |
| Frontend | React + TypeScript | Broad familiarity, strong ecosystem |
| Routing | TanStack Router | Type-safe, URL-first, good for namespace/pod params |
| Data fetching | React Query | Caching, loading states, refetch intervals |
| Frontend build | Vite | Fast HMR, straightforward TS/React config |
| Containerisation | Docker multi-stage | Small images, reproducible builds |
| Runtime base | `gcr.io/distroless/static` | Minimal attack surface, no shell |

---

## Repository structure

```
kube-portal/
├── api/
│   ├── cmd/server/
│   │   └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   │   ├── handlers.go
│   │   │   ├── namespaces.go
│   │   │   ├── deployments.go
│   │   │   ├── pods.go
│   │   │   ├── logs.go
│   │   │   ├── events.go
│   │   │   └── metrics.go
│   │   ├── k8s/
│   │   |   └── client.go
│   │   └── types/
│   │       └── k8s.go
│   ├── go.mod
│   └── go.sum
├── web/
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Workloads.tsx
│   │   │   ├── PodDetail.tsx
│   │   │   ├── Logs.tsx
│   │   │   └── Metrics.tsx
│   │   ├── components/
│   │   │   ├── NamespaceSelector.tsx
│   │   │   ├── LogViewer.tsx
│   │   │   └── EventsFeed.tsx
│   │   └── api/
│   │       └── client.ts
│   ├── package.json
│   └── vite.config.ts
├── deploy/
│   └── manifests/
│       ├── namespace.yaml
│       ├── serviceaccount.yaml
│       ├── clusterrole.yaml
│       ├── clusterrolebinding.yaml
│       ├── deployment.yaml
│       └── service.yaml
├── Dockerfile
├── Makefile
└── README.md
```

---

## Definition of done (v1)

- [ ] `make dev` starts a working out-of-cluster dev environment against Minikube
- [ ] `make deploy-local` builds, loads into Minikube, and deploys in-cluster successfully
- [ ] Can list and inspect deployments and pods across namespaces
- [ ] Can view last N lines of logs from any pod; graceful message if pod has no logs
- [ ] Link to external log platform shown in log view (if configured)
- [ ] Can view Kubernetes events, namespace-scoped
- [ ] Pod metrics shown where Metrics Server is available; graceful fallback where it isn't
- [ ] `kubectl apply -f deploy/manifests/` deploys the portal to a cluster with no manual steps
- [ ] Portal pod runs non-root with a read-only filesystem
- [ ] No raw Kubernetes API types exposed directly to the frontend
- [ ] README covers local dev and in-cluster deployment

---

## Guiding principles for the build

1. **One concern per package.** The k8s client lives in `internal/k8s`, handlers in `internal/handlers`. Don't mix them.
2. **Errors are first-class.** Every handler returns a typed error. Every UI state accounts for loading, error, and empty — not just the happy path.
3. **No raw k8s types on the wire.** Map to our own response types. This is the most important seam in the whole system.
4. **URLs are state.** Selected namespace, pod name, log cursor — all in the URL. Deep-linkable from day one.
5. **Degrade gracefully.** If Metrics Server is absent, say so clearly. If a pod has no logs, say so. Never a blank screen.