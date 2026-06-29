import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { WorkAddForm } from "./WorksSection";

describe("WorkAddForm", () => {
  it("ชื่อผลงานว่าง → ไม่เรียก onAdd และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<WorkAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มผลงาน" }));

    expect(await screen.findByText("กรุณากรอกชื่อผลงาน")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("กรอกชื่อผลงาน → เรียก onAdd พร้อมชื่อที่ตัดช่องว่างแล้ว", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<WorkAddForm onAdd={onAdd} submitting={false} />);

    await user.type(
      screen.getByPlaceholderText("เช่น รางวัลครูดีเด่น, ผลงานวิจัยในชั้นเรียน"),
      "  รางวัลครูดีเด่น  ",
    );
    await user.click(screen.getByRole("button", { name: "เพิ่มผลงาน" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ title: "รางวัลครูดีเด่น" }),
    );
  });
});
