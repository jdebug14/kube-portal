import { screen } from "@testing-library/react";
import { test, expect } from "vitest";
import { http, HttpResponse, delay } from "msw";
import { server } from "../test/server.ts";
import { renderWithQueryClient } from "../test/render.tsx";
import PodLogsViewer from "./PodLogsViewer.tsx";

test("happy path", async () => {
  let capturedUrl: string | undefined;
  server.use(
    http.get(
      "/api/v1/namespaces/test-namespace-1/pods/workload-1/logs",
      ({ request }) => {
        capturedUrl = request.url;
        return HttpResponse.text("Hello from server logs");
      },
    ),
  );
  renderWithQueryClient(
    <PodLogsViewer
      namespace="test-namespace-1"
      podName="workload-1"
      containers={["container-1", "container-2"]}
    />,
  );

  expect(await screen.findByText("Hello from server logs")).toBeInTheDocument();
  expect(capturedUrl).toContain("&container=container-1");
});

test("without containers", async () => {
  let capturedUrl: string | undefined;
  server.use(
    http.get(
      "/api/v1/namespaces/test-namespace-1/pods/workload-1/logs",
      ({ request }) => {
        capturedUrl = request.url;
        return HttpResponse.text("Hello from server logs");
      },
    ),
  );
  renderWithQueryClient(
    <PodLogsViewer namespace="test-namespace-1" podName="workload-1" />,
  );

  expect(await screen.findByText("Hello from server logs")).toBeInTheDocument();
  expect(capturedUrl).not.toContain("&container=");
});

test("empty response", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/pods/workload-1/logs", () => {
      return HttpResponse.text("");
    }),
  );

  renderWithQueryClient(
    <PodLogsViewer namespace="test-namespace-1" podName="workload-1" />,
  );

  expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument();
});

test("error state", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/pods/workload-1/logs", () => {
      return HttpResponse.json({ error: "pod not found" }, { status: 404 });
    }),
  );

  renderWithQueryClient(
    <PodLogsViewer namespace="test-namespace-1" podName="workload-1" />,
  );

  expect(await screen.findByText("Error: pod not found")).toBeInTheDocument();
});

test(
  "loading state",
  {
    retry: 2 /* some inherant flakiness using an artifical delay to test behavior*/,
  },
  async () => {
    server.use(
      http.get(
        "/api/v1/namespaces/test-namespace-1/pods/workload-1/logs",
        async () => {
          await delay(150); // small artificial delay so we can catch the loading state
          return HttpResponse.text("");
        },
      ),
    );
    renderWithQueryClient(
      <PodLogsViewer namespace="test-namespace-1" podName="workload-1" />,
    );

    expect(screen.queryByText(/Nothing to see here/)).toBeNull();
    expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument(); // confirms it eventually resolves
  },
);
