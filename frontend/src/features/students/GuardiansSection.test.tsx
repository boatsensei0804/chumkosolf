import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { GuardianInlineForm } from "./GuardiansSection";

describe("GuardianInlineForm", () => {
  it("กรอกไม่ครบ → error ไทย ไม่เรียก onAdd", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<GuardianInlineForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มผู้ปกครอง" }));

    expect(await screen.findByText("เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกชื่อ")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onAdd พร้อมข้อมูลผู้ปกครอง + ความสัมพันธ์", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<GuardianInlineForm onAdd={onAdd} submitting={false} />);

    await user.type(screen.getByLabelText("เลขบัตรประชาชน"), "1234567890123");
    await user.type(screen.getByLabelText("ชื่อ"), "สมพงษ์");
    await user.type(screen.getByLabelText("นามสกุล"), "ใจดี");
    await user.click(screen.getByRole("checkbox", { name: "ผู้ปกครองหลัก" }));
    await user.click(screen.getByRole("button", { name: "เพิ่มผู้ปกครอง" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({
        national_id: "1234567890123",
        first_name: "สมพงษ์",
        last_name: "ใจดี",
        relationship: "father",
        is_primary: true,
      }),
    );
  });
});
