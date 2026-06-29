import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { YearAddForm } from "./AcademicManager";

describe("YearAddForm", () => {
  it("ปีถูกต้อง (default พ.ศ. ปัจจุบัน) → เรียก onAdd พร้อมปี", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<YearAddForm onAdd={onAdd} submitting={false} />);

    await user.click(screen.getByRole("button", { name: "เพิ่มปีการศึกษา" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(expect.any(Number));
  });

  it("ปีไม่ถูกต้อง → ไม่เรียก onAdd และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(<YearAddForm onAdd={onAdd} submitting={false} />);

    const input = screen.getByRole("spinbutton");
    await user.clear(input);
    await user.type(input, "1990");
    await user.click(screen.getByRole("button", { name: "เพิ่มปีการศึกษา" }));

    expect(await screen.findByText("กรุณากรอกปีการศึกษา (พ.ศ.) ให้ถูกต้อง")).toBeInTheDocument();
    expect(onAdd).not.toHaveBeenCalled();
  });
});
