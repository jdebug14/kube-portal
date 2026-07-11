import { beforeEach, expect, test, vi } from "vitest";
import { apiFetch } from "./client";

const mockFetch = vi.fn();
const mockParse = vi.fn();

beforeEach(async () => {
  vi.stubGlobal("fetch", mockFetch);
  vi.resetAllMocks();
});

test("successful fetch call", async () => {
  // arrange
  const fakeUrl = "/some/fake/url";
  const fakeOkRes = { ok: true, json: async () => ({ fake: "data" }) };
  mockFetch.mockResolvedValueOnce(fakeOkRes);
  mockParse.mockReturnValueOnce("some parsed json");

  // act
  const result = await apiFetch(fakeUrl, mockParse);

  // assert
  expect(mockFetch).toHaveBeenCalledWith(fakeUrl);
  expect(mockParse).toHaveBeenCalledWith(fakeOkRes);
  expect(result).toBe("some parsed json");
});

test("failed fetch call", async () => {
  // arrange
  const fakeUrl = "/some/fake/url";
  const fakeErrorRes = {
    ok: false,
    json: async () => ({ error: "some error" }),
  };
  mockFetch.mockResolvedValueOnce(fakeErrorRes);

  // act & assert
  await expect(apiFetch(fakeUrl, mockParse)).rejects.toThrow(
    new Error("some error"),
  );
  expect(mockFetch).toHaveBeenCalledWith(fakeUrl);
  expect(mockParse).toHaveBeenCalledTimes(0);
});
