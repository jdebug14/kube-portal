import "@testing-library/jest-dom/vitest";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./test/server.ts";
import { cleanup } from "@testing-library/react";

// silence jsdom warnings
window.scrollTo = () => {};
// end silence warnings

beforeAll(() => {
  server.listen();
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
});

afterAll(() => {
  server.close();
});
