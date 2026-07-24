import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import QueryStatus from "./QueryStatus";

test("shows loading indicator when isLoading is true", () => {
  render(
    <QueryStatus
      isLoading={true}
      isLoadingError={false}
      isRefetchError={false}
      error={null}
    />,
  );

  expect(screen.getByRole("status")).toBeInTheDocument();
});

test("shows error notice when isLoadingError is true", () => {
  render(
    <QueryStatus
      isLoading={false}
      isLoadingError={true}
      isRefetchError={false}
      error={new Error("service unavailable")}
    />,
  );

  expect(screen.getByText(/Error: service unavailable/)).toBeInTheDocument();
});

test("shows stale-data notice when isRefetchError is true", () => {
  render(
    <QueryStatus
      isLoading={false}
      isLoadingError={false}
      isRefetchError={true}
      error={new Error("connection reset")}
    />,
  );

  expect(
    screen.getByText(/Refresh failed - showing last known data/),
  ).toBeInTheDocument();
});

test("renders nothing when all states are false", () => {
  const { container } = render(
    <QueryStatus
      isLoading={false}
      isLoadingError={false}
      isRefetchError={false}
      error={null}
    />,
  );

  expect(container).toBeEmptyDOMElement();
});
