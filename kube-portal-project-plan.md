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
- Live pod log streaming (WebSocket)
- Kubernetes Events feed
- Pod CPU and memory metrics via Metrics Server (graceful degradation if absent)
- Kubernetes manifests: Deployment, Service, ServiceAccount, ClusterRole, ClusterRoleBinding
- Single Docker image (Go binary + embedded React assets)

### Explicitly out of scope (v1)
| Item | Reason | Future milestone |
|---|---|---|
| Helm chart | Packaging, not functionality | v2 |
| OIDC / SSO auth | Mini-project on its own; use token passthrough for now | v2 |
| Restart / scale operations | Mutations need more hardening and RBAC care | v2 |
| Multi-cluster support | Complexity multiplier; nail single-cluster first | v3 |
| Prometheus / custom metrics | No external dependencies in v1 | v2 |
| Dark mode / theming | UI polish | v2 |

---

## Timeline

### Week 1 — Skeleton + Workloads

**Goal:** A working dev loop and the first real data on screen.

#### Day 1–2: Repo scaffold
- `git init kube-portal` with `.gitignore`, `README.md`, `Makefile`
- Go module: `api/` with `cmd/server/main.go`, `/healthz` endpoint
- React scaffold: `web/` with Vite + TypeScript + TanStack Router + React Query
- `make dev` runs both API and frontend with hot reload
- `deploy/manifests/` skeleton: `namespace.yaml`, `serviceaccount.yaml`

**Design decision to consider:** _How do we serve the frontend in production?_ Options: (a) embed React build into Go binary with `embed.FS`, (b) separate nginx container. We'll go with (a) — single binary is simpler to deploy on any cluster, no sidecar to manage.

#### Day 3–4: Kubernetes client + Workloads API
- `internal/k8s/client.go` — in-cluster vs kubeconfig detection, single shared client
- Handlers: `GET /api/v1/namespaces`, `GET /api/v1/namespaces/:ns/deployments`, `GET /api/v1/namespaces/:ns/pods`
- Typed response structs — no raw k8s objects leaked to the frontend
- RBAC: `ClusterRole` with `get/list/watch` on `namespaces`, `deployments`, `replicasets`, `pods`

**Design decision to consider:** _Typed response structs vs. forwarding raw k8s JSON._ Forwarding is faster to write but tightly couples the frontend to the k8s API shape. Typed structs give us a stable contract and let us evolve independently. We'll use typed structs with explicit field mapping.

#### Day 5: Workloads UI
- Namespace selector (global, persisted in URL param)
- Deployments list: name, namespace, ready/desired, age, image
- Pods list: name, status, restarts, node, age
- Pod detail page: labels, annotations, conditions, container statuses

**Checkpoint:** Can browse namespaces, see deployments and pods, click into a pod.

---

### Week 2 — Observability

**Goal:** A portal that gives you real signal — logs and events — the two things you reach for first in any incident.

#### Day 1–2: Live log streaming
- `internal/ws/logs.go` — WebSocket handler wrapping `client-go` pod log follow
- Reconnection logic: client-side exponential backoff, server-side stream cleanup on disconnect
- Timestamp toggle, line buffer cap (prevent memory growth on noisy pods)
- Frontend: `LogViewer` component with auto-scroll, pause-on-hover, line count

**Design decision to consider:** _WebSocket vs. Server-Sent Events (SSE) for log streaming._ SSE is simpler (HTTP, no upgrade, works through more proxies) but is unidirectional. WebSocket lets us send control signals (filter, pause) from client. We'll use WebSocket — the control channel will matter later.

#### Day 3: Events feed
- `GET /api/v1/namespaces/:ns/events` — filtered by involved object where applicable
- Warning events surfaced prominently (colour-coded in UI)
- Sorted by last timestamp descending
- Pod detail page shows related events inline

#### Day 4–5: Pod metrics
- `GET /api/v1/namespaces/:ns/pods/:pod/metrics` — via `metrics.k8s.io/v1beta1`
- Graceful degradation: if Metrics Server absent, API returns `metrics_available: false` and UI shows a callout instead of an error
- CPU (millicores) and memory (MiB) with request/limit context where available

**Checkpoint:** Can stream logs from any pod, see events, see metrics (or a clear "Metrics Server not available" state).

---

### Week 3 — Hardening + Deployable Artefact

**Goal:** Something you'd actually hand to a team to run. Not pretty, but trustworthy.

#### Day 1–2: Error handling + resilience
- Consistent API error envelope: `{ error: string, code: number }`
- Frontend error boundaries on every data-fetching page
- Loading skeletons (not spinners) — better perceived performance
- Handle common failure modes: namespace not found, pod already terminated, metrics unavailable

#### Day 3: RBAC + security posture
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
│   │   │   ├── namespaces.go
│   │   │   ├── workloads.go
│   │   │   ├── events.go
│   │   │   └── metrics.go
│   │   ├── k8s/
│   │   │   └── client.go
│   │   └── ws/
│   │       └── logs.go
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

- [ ] `make dev` starts a working local dev environment
- [ ] Can list and inspect deployments and pods across namespaces
- [ ] Can stream live logs from any running pod
- [ ] Can view Kubernetes events, namespace-scoped
- [ ] Pod metrics shown where Metrics Server is available; graceful fallback where it isn't
- [ ] `kubectl apply -f deploy/manifests/` deploys the portal to a cluster with no manual steps
- [ ] Portal pod runs non-root with a read-only filesystem
- [ ] No raw Kubernetes API types exposed directly to the frontend
- [ ] README covers local dev and in-cluster deployment

---

## Guiding principles for the build

1. **One concern per package.** The k8s client lives in `internal/k8s`, handlers in `internal/handlers`, WebSocket logic in `internal/ws`. Don't mix them.
2. **Errors are first-class.** Every handler returns a typed error. Every UI state accounts for loading, error, and empty — not just the happy path.
3. **No raw k8s types on the wire.** Map to our own response types. This is the most important seam in the whole system.
4. **URLs are state.** Selected namespace, pod name, log cursor — all in the URL. Deep-linkable from day one.
5. **Degrade gracefully.** If Metrics Server is absent, say so clearly. If a pod has no logs, say so. Never a blank screen.
