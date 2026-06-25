import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiRequestError } from "@/lib/api/client";

import { login } from "./api";

function mockFetchOnce(status: number, payload: unknown): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({
      status,
      json: () => Promise.resolve(payload),
    } as Response),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("login api", () => {
  it("ส่ง payload เป็น snake_case และคืน data ที่ผ่าน validate", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      status: 200,
      json: () =>
        Promise.resolve({
          success: true,
          error: null,
          data: {
            access_token: "a",
            refresh_token: "r",
            token_type: "Bearer",
            expires_in: 900,
            user: {
              id: "u1",
              username: "superadmin",
              role: "super_admin",
              school_id: "s1",
              is_school_admin: true,
              work_groups: [],
            },
          },
        }),
    } as Response);
    vi.stubGlobal("fetch", fetchMock);

    const result = await login({
      schoolCode: "CHUMKO",
      username: "superadmin",
      password: "admin1234",
    });

    expect(result.user.username).toBe("superadmin");

    // ตรวจว่า body ที่ยิงไปเป็น snake_case ตาม contract ของ backend
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(JSON.parse(init.body as string)).toEqual({
      school_code: "CHUMKO",
      username: "superadmin",
      password: "admin1234",
    });
  });

  it("backend ตอบ error → โยน ApiRequestError พร้อม code + ข้อความไทย", async () => {
    mockFetchOnce(401, {
      success: false,
      data: null,
      error: { code: "INVALID_CREDENTIALS", message: "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง" },
    });

    await expect(
      login({ schoolCode: "CHUMKO", username: "x", password: "y" }),
    ).rejects.toMatchObject({
      code: "INVALID_CREDENTIALS",
      message: "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง",
    });
    await expect(
      login({ schoolCode: "CHUMKO", username: "x", password: "y" }),
    ).rejects.toBeInstanceOf(ApiRequestError);
  });
});
