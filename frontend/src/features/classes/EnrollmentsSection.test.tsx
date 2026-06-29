import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { EnrollAddForm } from "./EnrollmentsSection";
import type { StudentListItem } from "@/shared/schemas/student";

function student(id: string, code: string, first: string): StudentListItem {
  return {
    id,
    student_code: code,
    status: "studying",
    prefix: "เด็กชาย",
    first_name: first,
    last_name: "ใจดี",
    national_id_masked: "",
    phone: "",
    created_at: "",
  };
}

describe("EnrollAddForm (bulk)", () => {
  it("แสดงข้อความเมื่อไม่มีนักเรียนให้เลือก", () => {
    render(<EnrollAddForm students={[]} onAdd={vi.fn()} submitting={false} />);
    expect(screen.getByText(/จัดนักเรียนครบแล้ว/)).toBeInTheDocument();
  });

  it("เลือกทั้งหมด → จัดเข้าห้อง เรียก onAdd พร้อม id ทุกคน", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <EnrollAddForm
        students={[student("s1", "S001", "ก"), student("s2", "S002", "ข")]}
        onAdd={onAdd}
        submitting={false}
      />,
    );

    await user.click(screen.getByRole("button", { name: /เลือกทั้งหมด/ }));
    await user.click(screen.getByRole("button", { name: /จัดเข้าห้อง/ }));

    await waitFor(() => expect(onAdd).toHaveBeenCalledTimes(1));
    expect(onAdd).toHaveBeenCalledWith(["s1", "s2"]);
  });
});
