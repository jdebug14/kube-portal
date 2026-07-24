import { screen } from "@testing-library/react";
import { test, expect } from "vitest";
import { http, HttpResponse, delay } from "msw";
import { renderWithQueryClient } from "../test/render.tsx";
import { server } from "../test/server.ts";
import EventsFeed from "./EventsFeed.tsx";

test("namespace events", async () => {
  let capturedUrl: string | undefined;
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/events", ({ request }) => {
      capturedUrl = request.url;
      return HttpResponse.json([
        {
          type: "Normal",
          reason: "Pulled",
          message:
            'Container image "my.test/image:1.0" already present on machine',
          count: 1,
          first_time: "2026-01-01T12:01:00-7:00",
          last_time: "2026-01-01T12:01:00-7:00",
          involved_object: {
            kind: "Pod",
            name: "workload-1",
            namespace: "test-namespace-1",
          },
        },
        {
          type: "Warning",
          reason: "Unhealthy",
          message: "Readiness probe failed",
          count: 3,
          first_time: "2026-01-01T12:01:00-7:00",
          last_time: "2026-01-01T12:03:00-7:00",
          involved_object: {
            kind: "Pod",
            name: "workload-2",
            namespace: "test-namespace-1",
          },
        },
      ]);
    }),
  );
  renderWithQueryClient(<EventsFeed namespace="test-namespace-1" />);

  expect(
    await screen.findByText(/name=workload-1.*reason=Pulled/),
  ).toBeInTheDocument();
  expect(
    await screen.findByText(/name=workload-2.*reason=Unhealthy/),
  ).toBeInTheDocument();
  expect(capturedUrl).not.toContain("?involvedObjectName=");
});

test("workload events", async () => {
  let capturedUrl: string | undefined;
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/events", ({ request }) => {
      capturedUrl = request.url;
      return HttpResponse.json([
        {
          type: "Normal",
          reason: "Pulled",
          message:
            'Container image "my.test/image:1.0" already present on machine',
          count: 1,
          first_time: "2026-01-01T12:01:00-7:00",
          last_time: "2026-01-01T12:01:00-7:00",
          involved_object: {
            kind: "Pod",
            name: "workload-1",
            namespace: "test-namespace-1",
          },
        },
      ]);
    }),
  );

  renderWithQueryClient(
    <EventsFeed namespace="test-namespace-1" involvedObjectName="workload-1" />,
  );

  expect(
    await screen.findByText(/name=workload-1.*reason=Pulled/),
  ).toBeInTheDocument();
  expect(capturedUrl).toContain("?involvedObjectName=workload-1");
});

test("empty", async () => {
  renderWithQueryClient(<EventsFeed namespace="test-namespace-2" />);

  expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument();
});

test("shows error state", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/events", () => {
      return HttpResponse.json(
        { error: "service unavailable" },
        { status: 500 },
      );
    }),
  );
  renderWithQueryClient(<EventsFeed namespace="test-namespace-1" />);

  expect(
    await screen.findByText("Error: service unavailable"),
  ).toBeInTheDocument();
});

test(
  "loading state",
  {
    retry: 2 /* some inherant flakiness using an artifical delay to test behavior*/,
  },
  async () => {
    server.use(
      http.get("/api/v1/namespaces/test-namespace-1/events", async () => {
        await delay(150); // small artificial delay so we can catch the loading state
        return HttpResponse.json([]);
      }),
    );
    renderWithQueryClient(<EventsFeed namespace="test-namespace-1" />);

    expect(screen.queryByText(/Nothing to see here/)).toBeNull();
    expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument(); // confirms it eventually resolves
  },
);
