import { screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { renderWithRouter } from "../test/render.tsx";
import userEvent from "@testing-library/user-event";
import { server } from "../test/server.ts";
import { http, HttpResponse, delay } from "msw";

vi.mock("../components/Pods", () => ({ default: () => null }));
vi.mock("../components/EventsFeed", () => ({ default: () => null }));
const user = userEvent.setup();

test("happy path", async () => {
  renderWithRouter("/namespaces/test-namespace-1");

  expect(
    await screen.findByRole("heading", {
      name: "Deployments",
      level: 3,
    }),
  ).toBeInTheDocument();

  expect(await screen.findByText("deployment-1")).toBeInTheDocument();
  expect(await screen.findByText("deployment-2")).toBeInTheDocument();
  expect(await screen.findByText("deployment-3")).toBeInTheDocument();
});

test("empty response", async () => {
  renderWithRouter("/namespaces/test-namespace-2");

  expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument();
});

test("filterted deployments", async () => {
  renderWithRouter("/namespaces/test-namespace-1");

  expect(
    await screen.findByRole("heading", {
      name: "Deployments",
      level: 3,
    }),
  ).toBeInTheDocument();

  const searchBox = await screen.findByPlaceholderText("Type to search...");
  await user.type(searchBox, "1");
  expect(await screen.findByText("deployment-1")).toBeInTheDocument();
  expect(screen.queryByText("deployment-2")).toBeNull();
  expect(screen.queryByText("deployment-3")).toBeNull();

  await user.clear(searchBox);
  await user.type(searchBox, "deplo");
  expect(await screen.findByText("deployment-1")).toBeInTheDocument();
  expect(await screen.findByText("deployment-2")).toBeInTheDocument();
  expect(await screen.findByText("deployment-3")).toBeInTheDocument();

  await user.clear(searchBox);
  await user.type(searchBox, "works");
  expect(screen.getByText(/Nothing to see here/)).toBeInTheDocument();
});

test("error state", async () => {
  server.use(
    http.get("/api/v1/namespaces/test-namespace-1/deployments", () => {
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
      http.get("/api/v1/namespaces/test-namespace-1/deployments", async () => {
        await delay(150); // small artificial delay so we can catch the loading state
        return HttpResponse.json([]);
      }),
    );
    renderWithRouter("/namespaces/test-namespace-1");

    expect(screen.queryByText(/Nothing to see here/)).toBeNull();
    expect(await screen.findByText(/Nothing to see here/)).toBeInTheDocument(); // confirms it eventually resolves
  },
);
