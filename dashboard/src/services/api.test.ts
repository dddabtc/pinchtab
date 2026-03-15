import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fetchTabScreenshot } from "./api";

describe("fetchTabScreenshot", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    window.localStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("requests the existing tab screenshot endpoint and returns a blob", async () => {
    const responseBlob = new Blob(["image-bytes"], { type: "image/png" });
    fetchMock.mockResolvedValue({
      ok: true,
      blob: vi.fn().mockResolvedValue(responseBlob),
    });
    window.localStorage.setItem("pinchtab.auth.token", "secret-token");

    const result = await fetchTabScreenshot("tab_123");

    expect(result).toBeInstanceOf(Blob);
    expect(await result.text()).toBe("image-bytes");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0]?.[0]).toBe("/tabs/tab_123/screenshot");
    expect(
      new Headers(fetchMock.mock.calls[0]?.[1]?.headers).get("Authorization"),
    ).toBe("Bearer secret-token");
  });
});
