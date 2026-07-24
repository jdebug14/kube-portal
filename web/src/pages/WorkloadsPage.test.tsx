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
  expect(screen.getByText(/Nothing to see here/)).toBeInTheDocument();
});

test("empty response", async () => {
  renderWithRouter("/namespaces/test-namespace-2");

  expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument();
});

test("error state", async () => {
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
  "loading state",
  {
    retry: 2 /* some inherant flakiness using an artifical delay to test behavior*/,
  },
  async () => {
    server.use(
      http.get("/api/v1/namespaces/test-namespace-1/pods", async () => {
        await delay(150); // small artificial delay so we can catch the loading state
        return HttpResponse.json([]);
      }),
    );
    renderWithRouter("/namespaces/test-namespace-1");

    expect(screen.queryByText(/Nothing to see here/)).toBeNull();
    expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument(); // confirms it eventually resolves
  },
);
