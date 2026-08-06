import { screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { renderWithRouter } from "../test/render.tsx";
import { server } from "../test/server.ts";
import { http, HttpResponse, delay } from "msw";

vi.mock("../components/EventsFeed", () => ({ default: () => null }));

test("happy path", async () => {
  renderWithRouter("/namespaces/test-namespace-1/deployments/deployment-1");

  expect(
    await screen.findByRole("heading", {
      name: "Deployment: deployment-1",
      level: 2,
    }),
  ).toBeInTheDocument();

  expect(await screen.findByText("RollingUpdate")).toBeInTheDocument();
  expect(
    await screen.findByText(/3 desired, 3 updated, 2 ready, 2 available/),
  ).toBeInTheDocument();
  expect(await screen.findByText("test-namespace-1")).toBeInTheDocument();
  expect(
    await screen.findByText("my.test/annotation=true"),
  ).toBeInTheDocument();
  expect(
    await screen.findByText("app.kubernetes.io/name=my-app"),
  ).toBeInTheDocument();
  expect(await screen.findByText(/Available: False/)).toBeInTheDocument();
  expect(
    await screen.findByText(/Deployment does not have minimum availability/),
  ).toBeInTheDocument();

  expect(await screen.findByText("3/2")).toBeInTheDocument();
  expect(await screen.findByText(/Pod template:/)).toBeInTheDocument();
  expect(
    await screen.findByText("Image: my.test/web:v1.0"),
  ).toBeInTheDocument();
});

test("error state", async () => {
  server.use(
    http.get(
      "/api/v1/namespaces/test-namespace-1/deployments/deployment-1",
      () => {
        return HttpResponse.json(
          { error: "service unavailable" },
          { status: 500 },
        );
      },
    ),
  );

  renderWithRouter("/namespaces/test-namespace-1/deployments/deployment-1");

  expect(
    await screen.findByText("Error: service unavailable"),
  ).toBeInTheDocument();
});

test(
  "loading state",
  {
    retry: 2,
  },
  async () => {
    server.use(
      http.get(
        "/api/v1/namespaces/test-namespace-1/deployments/deployment-1",
        async () => {
          await delay(150);
          return HttpResponse.json({
            name: "deployment-1",
            namespace: "test-namespace-1",
            strategy: "Recreate",
            desired_replicas: 1,
            updated_replicas: 1,
            ready_replicas: 1,
            available_replicas: 1,
            rollout_status: "Complete",
            conditions: [],
            revisions: [],
            created_at: "2026-01-01T00:00:10-07:00",
          });
        },
      ),
    );

    renderWithRouter("/namespaces/test-namespace-1/deployments/deployment-1");

    expect(screen.queryByText("Deployment: deployment-1")).toBeNull();
    expect(
      await screen.findByRole("status", { hidden: true }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole("heading", {
        name: "Deployment: deployment-1",
        level: 2,
      }),
    ).toBeInTheDocument();
  },
);
