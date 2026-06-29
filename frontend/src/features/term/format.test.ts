import { describe, expect, it } from "vitest";

import { formatTerm } from "./format";

describe("formatTerm", () => {
  it("มีเทอมปัจจุบัน → แสดงปีการศึกษา + ภาคเรียน", () => {
    expect(formatTerm({ has_current: true, academic_year: 2568, term: 1 })).toBe(
      "ปีการศึกษา 2568 · ภาคเรียนที่ 1",
    );
  });

  it("ยังไม่กำหนดเทอม → แสดงข้อความเตือน", () => {
    expect(formatTerm({ has_current: false, academic_year: 0, term: 0 })).toBe(
      "ยังไม่ได้กำหนดเทอมปัจจุบัน",
    );
  });
});
