import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import Notice from "./Notice";

test("renders info notice with icon and message", () => {
  render(<Notice type="info">Something informational</Notice>);

  expect(screen.getByText("ℹ️")).toBeInTheDocument();
  expect(screen.getByText("Something informational")).toBeInTheDocument();
});

test("renders warning notice with icon and message", () => {
  render(<Notice type="warning">Something concerning</Notice>);

  expect(screen.getByText("⚠️")).toBeInTheDocument();
  expect(screen.getByText("Something concerning")).toBeInTheDocument();
});

test("renders error notice with icon and message", () => {
  render(<Notice type="error">Something broke</Notice>);

  expect(screen.getByText("❌")).toBeInTheDocument();
  expect(screen.getByText("Something broke")).toBeInTheDocument();
});
