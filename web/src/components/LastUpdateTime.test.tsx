import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import LastUpdateTime from "./LastUpdateTime";

test("happy path", () => {
  render(<LastUpdateTime timestamp={100}></LastUpdateTime>);

  expect(screen.getByText(/Last updated/)).toBeInTheDocument();
});

test("zero timestamp", () => {
  const { container } = render(<LastUpdateTime timestamp={0}></LastUpdateTime>);

  expect(container).toBeEmptyDOMElement();
});
