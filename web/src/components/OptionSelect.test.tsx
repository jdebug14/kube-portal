import { render, screen, within } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import OptionSelect from "./OptionSelect";

const user = userEvent.setup();
const mockChangeHandler = vi.fn();

test("with strings", async () => {
  const testValue = "hello";
  render(
    <OptionSelect
      label="Select a word: "
      kind="string"
      value={testValue}
      changeHandler={mockChangeHandler}
      options={[
        ["hello", "hello"],
        ["world", "world"],
      ]}
    />,
  );

  const selectEl = screen.getByRole("combobox", { name: /Select a word:/ });
  await user.selectOptions(selectEl, "world");
  const options = within(selectEl).queryAllByRole("option");
  expect(options).toHaveLength(2);
  expect(mockChangeHandler).toHaveBeenCalledWith("world");
});

test("with numbers", async () => {
  const testValue = 1;
  render(
    <OptionSelect
      label="Select a number: "
      kind="number"
      value={testValue}
      changeHandler={mockChangeHandler}
      options={[
        ["One", 1],
        ["Two", 2],
      ]}
    />,
  );

  const selectEl = screen.getByRole("combobox", { name: /Select a number:/ });
  const options = within(selectEl).queryAllByRole("option");
  expect(options).toHaveLength(2);
  await user.selectOptions(selectEl, "Two");
  expect(mockChangeHandler).toHaveBeenCalledWith(2);
});

test("no options", async () => {
  const testValue = 1;
  render(
    <OptionSelect
      label="Select a number: "
      kind="number"
      value={testValue}
      changeHandler={mockChangeHandler}
      options={[]}
    />,
  );

  const selectEl = screen.getByRole("combobox", { name: /Select a number:/ });
  const options = within(selectEl).queryAllByRole("option");
  expect(options).toHaveLength(0);
});
