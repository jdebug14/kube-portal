import { screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { renderWithRouter } from "../test/render.tsx";
import userEvent from "@testing-library/user-event";
import { server } from "../test/server.ts";
import { http, HttpResponse, delay } from "msw";

const user = userEvent.setup();

test("happy path", async () => {
  renderWithRouter("/");

  expect(await screen.findByText("test-namespace-1")).toBeInTheDocument();
  expect(await screen.findByText("test-namespace-2")).toBeInTheDocument();
  expect(await screen.findByText("test-namespace-3")).toBeInTheDocument();

  const filterInput = screen.getByPlaceholderText("Type to search...");
  await user.type(filterInput, "1");
  expect(await screen.findByText("test-namespace-1")).toBeInTheDocument();
  expect(screen.queryByText("test-namespace-2")).toBeNull();
  expect(screen.queryByText("test-namespace-3")).toBeNull();

  await user.clear(filterInput);
  await user.type(filterInput, "namespa");
  expect(await screen.findByText("test-namespace-1")).toBeInTheDocument();
  expect(await screen.findByText("test-namespace-2")).toBeInTheDocument();
  expect(await screen.findByText("test-namespace-3")).toBeInTheDocument();

  await user.clear(filterInput);
  await user.type(filterInput, "namess");
  expect(screen.getByText("No namespaces to show.")).toBeInTheDocument();
});

test("no namespaces", async () => {
  server.use(
    http.get("/api/v1/namespaces", () => {
      return HttpResponse.json([]);
    }),
  );
  renderWithRouter("/");

  expect(await screen.findByText("No namespaces to show.")).toBeInTheDocument();
});

test("shows error state", async () => {
  server.use(
    http.get("/api/v1/namespaces", () => {
      return HttpResponse.json(
        { error: "service unavailable" },
        { status: 500 },
      );
    }),
  );
  renderWithRouter("/");

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
      http.get("/api/v1/namespaces", async () => {
        await delay(100); // small artificial delay so we can catch the loading state
        return HttpResponse.json([]);
      }),
    );
    renderWithRouter("/");

    expect(await screen.findByText("Loading...")).toBeInTheDocument();
    expect(
      await screen.findByText("No namespaces to show."),
    ).toBeInTheDocument(); // confirms it eventually resolves
  },
);
