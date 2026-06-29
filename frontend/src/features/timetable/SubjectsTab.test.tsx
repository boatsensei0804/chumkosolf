import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { SubjectAddForm } from "./SubjectsTab";

describe("SubjectAddForm", () => {
  it("ไม่กรอกรหัส/ชื่อวิชา → ไม่เรียก onAdd และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<SubjectAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มวิชา" }));

    expect(await screen.findByText("กรุณากรอกรหัสวิชาและชื่อวิชา")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("กรอกรหัส + ชื่อวิชา → เรียก onAdd พร้อมค่าที่ตัดช่องว่าง", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<SubjectAddForm onAdd={onAdd} submitting={false} />);

    await user.type(screen.getByPlaceholderText("เช่น ค21101"), "  ค21101  ");
    await user.type(screen.getByPlaceholderText("เช่น คณิตศาสตร์พื้นฐาน"), "คณิตศาสตร์");
    await user.click(screen.getByRole("button", { name: "เพิ่มวิชา" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ subject_code: "ค21101", name: "คณิตศาสตร์" }),
    );
  });
});
