import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import KeyValueList from "./KeyValueList";

test("happy path", () => {
  render(
    <KeyValueList
      title="Labels"
      entries={[
        ["some", "label"],
        ["hello", "world"],
      ]}
    ></KeyValueList>,
  );
  expect(screen.getByText("Labels:")).toBeInTheDocument();
  expect(screen.getByText("some: label")).toBeInTheDocument();
  expect(screen.getByText("hello: world")).toBeInTheDocument();
});

test("empty", () => {
  render(<KeyValueList title="Labels" entries={[]}></KeyValueList>);

  expect(screen.getByText("Labels:")).toBeInTheDocument();
  expect(screen.getByText("None")).toBeInTheDocument();
});
