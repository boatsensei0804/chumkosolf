import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { PositionAddForm, StandingAddForm } from "./PersonnelSubResources";

describe("PositionAddForm", () => {
  it("เพิ่มตำแหน่ง → เรียก onAdd พร้อม position (default รองผู้อำนวยการ)", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<PositionAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มตำแหน่ง" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ position: "deputy_director" }),
    );
  });
});

describe("StandingAddForm", () => {
  it("ชื่อวิทยฐานะว่าง → ไม่เรียก onAdd และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<StandingAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มวิทยฐานะ" }));

    expect(await screen.findByText("กรุณากรอกชื่อวิทยฐานะ")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("กรอกชื่อ + ติ๊กปัจจุบัน → เรียก onAdd พร้อมข้อมูล", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<StandingAddForm onAdd={onAdd} submitting={false} />);

    await user.type(screen.getByPlaceholderText("เช่น ครู คศ.1, ชำนาญการ"), "ชำนาญการ");
    await user.click(screen.getByRole("checkbox", { name: "ปัจจุบัน" }));
    await user.click(screen.getByRole("button", { name: "เพิ่มวิทยฐานะ" }));

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ standing: "ชำนาญการ", is_current: true }),
    );
  });
});
