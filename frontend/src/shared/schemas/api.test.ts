import { describe, expect, it } from "vitest";
import { z } from "zod";

import { apiResponseSchema } from "./api";
import { attendanceStatusSchema } from "./enums";

describe("apiResponseSchema", () => {
  const studentResponse = apiResponseSchema(z.object({ id: z.string() }));

  it("parse response สำเร็จที่มี data", () => {
    const parsed = studentResponse.parse({
      success: true,
      data: { id: "abc" },
      error: null,
    });
    expect(parsed.success).toBe(true);
    expect(parsed.data?.id).toBe("abc");
  });

  it("parse response error ที่ data เป็น null", () => {
    const parsed = studentResponse.parse({
      success: false,
      data: null,
      error: { code: "NOT_FOUND", message: "ไม่พบข้อมูล" },
    });
    expect(parsed.error?.code).toBe("NOT_FOUND");
  });
});

describe("attendanceStatusSchema", () => {
  it("รับค่าสถานะที่ถูกต้อง", () => {
    expect(attendanceStatusSchema.parse("present")).toBe("present");
  });

  it("ปฏิเสธค่าสถานะที่ไม่รู้จัก", () => {
    expect(() => attendanceStatusSchema.parse("unknown")).toThrow();
  });
});
