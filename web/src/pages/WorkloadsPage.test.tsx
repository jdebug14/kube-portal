import { screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { renderWithRouter } from "../test/render.tsx";
import userEvent from "@testing-library/user-event";
import { server } from "../test/server.ts";
import { http, HttpResponse, delay } from "msw";

const user = userEvent.setup();
vi.mock("../components/EventsFeed", () => ({ default: () => null }));

test("happy path", async () => {
  renderWithRouter("/namespaces/test-namespace-1");

  expect(await screen.findByText("workload-1")).toBeInTheDocument();
  expect(await screen.findByText("workload-2")).toBeInTheDocument();
  expect(await screen.findByText("workload-3")).toBeInTheDocument();

  const filterInput = screen.getByPlaceholderText("Type to search...");
  await user.type(filterInput, "1");
  expect(await screen.findByText("workload-1")).toBeInTheDocument();
  expect(screen.queryByText("workload-2")).toBeNull();
  expect(screen.queryByText("workload-3")).toBeNull();

  await user.clear(filterInput);
  await user.type(filterInput, "work");
  expect(await screen.findByText("workload-1")).toBeInTheDocument();
  expect(await screen.findByText("workload-2")).toBeInTheDocument();
  expect(await screen.findByText("workload-3")).toBeInTheDocument();

  await user.clear(filterInput);
  await user.type(filterInput, "works");
  expect(screen.getByText("No workloads to show.")).toBeInTheDocument();
});

test("no workloads", async () => {
  renderWithRouter("/namespaces/test-namespace-2");

  expect(await screen.findByText("No workloads to show.")).toBeInTheDocument();
});

test("shows error state", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/pods", () => {
      return HttpResponse.json(
        { error: "service unavailable" },
        { status: 500 },
      );
    }),
  );
  renderWithRouter("/namespaces/test-namespace-1");

  expect(
    await screen.findByText("Error: service unavailable"),
  ).toBeInTheDocument();
});

test(
  "shows loading state",
  {
    retry: 2 /* some inherant flakiness using an artifical delay to test behavior*/,
  },
  async () => {
    server.use(
      http.get("/api/v1/namespaces/test-namespace-1/pods", async () => {
        await delay(100); // small artificial delay so we can catch the loading state
        return HttpResponse.json([]);
      }),
    );
    renderWithRouter("/namespaces/test-namespace-1");

    expect(await screen.findByText("Loading...")).toBeInTheDocument();
    expect(
      await screen.findByText("No workloads to show."),
    ).toBeInTheDocument(); // confirms it eventually resolves
  },
);
