import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { WorkGroup } from "@/shared/schemas/workGroup";

import { WorkGroupAddForm } from "./WorkGroupsSection";

const groups: WorkGroup[] = [
  { id: "wg-personnel", code: "personnel", name: "กลุ่มงานบุคคล" },
  { id: "wg-academic", code: "academic", name: "กลุ่มงานวิชาการ" },
];

describe("WorkGroupAddForm", () => {
  it("มอบหมาย → เรียก onAdd พร้อมกลุ่มแรกเป็น default", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<WorkGroupAddForm groups={groups} onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "มอบหมาย" }));

    expect(onAdd).toHaveBeenCalledWith({ work_group_id: "wg-personnel", is_group_admin: false });
  });

  it("ติ๊กหัวหน้ากลุ่ม → is_group_admin = true", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<WorkGroupAddForm groups={groups} onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("checkbox", { name: "เป็นหัวหน้ากลุ่ม" }));
    await user.click(screen.getByRole("button", { name: "มอบหมาย" }));

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ is_group_admin: true }),
    );
  });

  it("ไม่มีกลุ่มว่างให้มอบหมาย → แสดงข้อความ ไม่มีปุ่ม", () => {
    render(<WorkGroupAddForm groups={[]} onAdd={vi.fn()} submitting={false} />);
    expect(screen.getByText("มอบหมายครบทุกกลุ่มงานแล้ว")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "มอบหมาย" })).not.toBeInTheDocument();
  });
});
