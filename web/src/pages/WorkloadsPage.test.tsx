import { screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { renderWithRouter } from "../test/render.tsx";

vi.mock("../components/Deployments", () => ({ default: () => null }));
vi.mock("../components/Pods", () => ({ default: () => null }));
vi.mock("../components/EventsFeed", () => ({ default: () => null }));

test("happy path", async () => {
  renderWithRouter("/namespaces/test-namespace-1");

  expect(
    await screen.findByRole("heading", {
      name: "Namespace: test-namespace-1",
      level: 2,
    }),
  ).toBeInTheDocument();
});
