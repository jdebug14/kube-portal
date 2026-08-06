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
        name: "pod-1",
        namespace: "test-namespace-1",
        phase: "Running",
        created_at: "2026-01-01T00:00:10-07:00",
      },
      {
        name: "pod-2",
        namespace: "test-namespace-1",
        phase: "Running",
        created_at: "2026-01-01T00:01:10-07:00",
      },
      {
        name: "pod-3",
        namespace: "test-namespace-1",
        phase: "Pending",
        created_at: "2026-01-02T00:00:10-07:00",
      },
    ]);
  }),
  http.get("/api/v1/namespaces/test-namespace-2/pods", () => {
    return HttpResponse.json([]);
  }),
  http.get("/api/v1/namespaces/test-namespace-1/pods/pod-1", () => {
    return HttpResponse.json({
      name: "pod-1",
      namespace: "test-namespace-1",
      phase: "Running",
      host_name: "node-1",
      created_at: "2026-01-01T00:00:10-07:00",
      annotations: {
        "my.test/annotation": "true",
        "my.test/owner": "johnwick@continental.com",
      },
      labels: {
        "my.test/label": "helloworld",
        "my.test/component": "data-processing",
        "my.test/tier": "producer",
      },
      containers: [
        {
          name: "mycontainer",
          image: "my.test/image:v1.0",
          ready: true,
          restarts: 3,
          last_exit_time: "2026-01-01T12:00:00-7:00",
          last_exit_reason: "Error",
          last_exit_message: "OOM killed",
        },
      ],
    });
  }),
  http.get("/api/v1/namespaces/test-namespace-1/deployments", () => {
    return HttpResponse.json([
      {
        name: "deployment-1",
        namespace: "test-namespace-1",
        desired_replicas: 1,
        ready_replicas: 1,
        available_replicas: 1,
        created_at: "2026-01-01T00:00:10-07:00",
      },
      {
        name: "deployment-2",
        namespace: "test-namespace-1",
        desired_replicas: 2,
        ready_replicas: 2,
        available_replicas: 2,
        created_at: "2026-01-01T00:01:10-07:00",
      },
      {
        name: "deployment-3",
        namespace: "test-namespace-1",
        desired_replicas: 5,
        ready_replicas: 4,
        available_replicas: 3,
        created_at: "2026-01-02T00:00:10-07:00",
      },
    ]);
  }),
  http.get(
    "/api/v1/namespaces/test-namespace-1/deployments/deployment-1",
    () => {
      return HttpResponse.json({
        name: "deployment-1",
        namespace: "test-namespace-1",
        strategy: "RollingUpdate",
        desired_replicas: 3,
        updated_replicas: 3,
        ready_replicas: 2,
        available_replicas: 2,
        rollout_status: "InProgress",
        annotations: {
          "my.test/annotation": "true",
        },
        labels: {
          "app.kubernetes.io/name": "my-app",
        },
        conditions: [
          {
            type: "Available",
            status: "False",
            reason: "MinimumReplicasUnavailable",
            message: "Deployment does not have minimum availability.",
            last_update_time: "2026-01-01T00:00:10-07:00",
            last_transition_time: "2026-01-01T00:00:10-07:00",
          },
        ],
        revisions: [
          {
            name: "deployment-1-0001",
            number: 1,
            is_current: false,
            desired_replicas: 3,
            current_replicas: 3,
            ready_replicas: 2,
            pod_template: {
              annotations: {
                "pod.test/annotation": "yes",
              },
              labels: {
                "pod.test/label": "frontend",
              },
              containers: [
                {
                  name: "web",
                  image: "my.test/web:v1.0",
                  command: ["/start"],
                  args: ["--port", "8080"],
                },
              ],
            },
            created_at: "2026-01-02T00:00:10-07:00",
          },
        ],
        created_at: "2026-01-01T00:00:10-07:00",
      });
    },
  ),
  http.get("/api/v1/namespaces/test-namespace-2/deployments", () => {
    return HttpResponse.json([]);
  }),
];
