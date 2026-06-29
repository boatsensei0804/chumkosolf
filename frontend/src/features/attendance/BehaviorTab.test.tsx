import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { BehaviorAddForm } from "./BehaviorTab";

describe("BehaviorAddForm", () => {
  it("ไม่กรอกเหตุผล → ไม่เรียก onAdd และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<BehaviorAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "บันทึก" }));

    expect(await screen.findByText("กรุณาระบุเหตุผล")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("หักคะแนน (default) + เหตุผล → เรียก onAdd พร้อม points ติดลบ", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<BehaviorAddForm onAdd={onAdd} submitting={false} />);

    await user.type(screen.getByPlaceholderText("เช่น มาสาย, ช่วยงานส่วนรวม"), "มาสาย");
    await user.click(screen.getByRole("button", { name: "บันทึก" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ points: -5, reason: "มาสาย" }),
    );
  });
});
