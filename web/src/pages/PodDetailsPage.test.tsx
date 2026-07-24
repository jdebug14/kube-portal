import { screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { renderWithRouter } from "../test/render.tsx";
import { server } from "../test/server.ts";
import { http, HttpResponse } from "msw";

vi.mock("../components/EventsFeed", () => ({ default: () => null }));
vi.mock("../components/PodLogsViewer", () => ({ default: () => null }));

test("happy path", async () => {
  renderWithRouter("/namespaces/test-namespace-1/pods/workload-1");

  expect(await screen.findByText("workload-1")).toBeInTheDocument();
  expect(await screen.findByText(/Status:/)).toBeInTheDocument();
  expect(await screen.findByText(/Host node:/)).toBeInTheDocument();
  expect(await screen.findByText(/Created at:/)).toBeInTheDocument();

  const annotationsText = await screen.findByText(/Annotations:/i);
  expect(annotationsText).toBeInTheDocument();
  const annotationsList = annotationsText.nextElementSibling;
  expect(annotationsList).toBeInTheDocument();
  const annotations = annotationsList?.querySelectorAll("li");
  expect(annotations?.length).toBe(2);

  const labelsText = await screen.findByText(/Labels:/i);
  expect(labelsText).toBeInTheDocument();
  const labelsList = labelsText.nextElementSibling;
  expect(labelsList).toBeInTheDocument();
  const labels = labelsList?.querySelectorAll("li");
  expect(labels?.length).toBe(3);

  const containersText = await screen.findByText(/Containers:/i);
  expect(containersText).toBeInTheDocument();
  const containersList = containersText.nextElementSibling;
  expect(containersList).toBeInTheDocument();
  const containers = containersList?.querySelectorAll("li");
  expect(containers?.length).toBe(1);
});

test("error state", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/pods/workload-1", () => {
      return HttpResponse.json(
        { error: "service unavailable" },
        { status: 500 },
      );
    }),
  );
  renderWithRouter("/namespaces/test-namespace-1/pods/workload-1");

  expect(
    await screen.findByText("Error: service unavailable"),
  ).toBeInTheDocument();
});
