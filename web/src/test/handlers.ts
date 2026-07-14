import { http, HttpResponse } from "msw";

// Individual tests can override these with server.use(...) for
// test-specific responses. Add default/shared handlers here as needed.
export const handlers = [
  http.get("/api/v1/namespaces", () => {
    return HttpResponse.json([
      {
        name: "test-namespace-1",
        status: "Active",
        created_at: "2026-01-01T00:00:10-07:00",
      },
      {
        name: "test-namespace-2",
        status: "Active",
        created_at: "2026-01-01T00:01:10-07:00",
      },
      {
        name: "test-namespace-3",
        status: "Terminating",
        created_at: "2026-01-02T00:00:10-07:00",
      },
    ]);
  }),
  http.get("/api/v1/namespaces/test-namespace-1/pods", () => {
    return HttpResponse.json([
      {
        name: "workload1",
        namespace: "test-namespace-1",
        phase: "Running",
        created_at: "2026-01-01T00:00:10-07:00",
      },
      {
        name: "workload2",
        namespace: "test-namespace-1",
        phase: "Running",
        created_at: "2026-01-01T00:01:10-07:00",
      },
      {
        name: "workload3",
        namespace: "test-namespace-1",
        phase: "Pending",
        created_at: "2026-01-02T00:00:10-07:00",
      },
    ]);
  }),
];
