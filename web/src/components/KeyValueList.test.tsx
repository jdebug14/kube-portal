import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import KeyValueList from "./KeyValueList";

test("happy path", () => {
  render(
    <KeyValueList
      entries={[
        ["some", "label"],
        ["hello", "world"],
      ]}
    ></KeyValueList>,
  );
  expect(screen.getByText("some=label")).toBeInTheDocument();
  expect(screen.getByText("hello=world")).toBeInTheDocument();
});

test("empty", () => {
  const { container } = render(<KeyValueList entries={[]}></KeyValueList>);
  expect(container).toBeEmptyDOMElement();
});
