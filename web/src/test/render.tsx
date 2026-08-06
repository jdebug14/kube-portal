import {
  createMemoryHistory,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, type RenderResult } from "@testing-library/react";
import { routeTree } from "../router";
import type { ReactElement } from "react";

function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

export function renderWithQueryClient(ui: ReactElement): RenderResult {
  return render(
    <QueryClientProvider client={createTestQueryClient()}>
      {ui}
    </QueryClientProvider>,
  );
}

export function renderWithRouter(path: string): RenderResult {
  const history = createMemoryHistory({ initialEntries: [path] });
  const router = createRouter({ routeTree, history });

  return renderWithQueryClient(<RouterProvider router={router} />);
}
