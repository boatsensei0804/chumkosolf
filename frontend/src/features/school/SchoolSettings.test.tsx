import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render as rtlRender, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { SchoolInfoForm } from "./SchoolSettings";
import type { School } from "@/shared/schemas/school";

function render(ui: ReactNode): ReturnType<typeof rtlRender> {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return rtlRender(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

const school: School = {
  id: "s1",
  name: "โรงเรียนชุมโค",
  code: "CHUMKO",
  address: { house_no: "", moo: "", road: "", subdistrict: "", district: "", province: "", postal_code: "" },
  phone: "",
  email: "",
  website: "",
  director_name: "",
  is_active: true,
  attendance_late_after: "08:00",
  attendance_late_penalty: 5,
};

describe("SchoolInfoForm", () => {
  it("ลบชื่อโรงเรียน → ไม่เรียก onSave และแสดง error ไทย", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<SchoolInfoForm school={school} submitting={false} onSave={onSave} />);

    await user.clear(screen.getByLabelText("ชื่อโรงเรียน"));
    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลโรงเรียน" }));

    expect(await screen.findByText("กรุณากรอกชื่อโรงเรียน")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("แก้ชื่อ → เรียก onSave พร้อมชื่อใหม่", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<SchoolInfoForm school={school} submitting={false} onSave={onSave} />);

    const name = screen.getByLabelText("ชื่อโรงเรียน");
    await user.clear(name);
    await user.type(name, "โรงเรียนใหม่");
    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลโรงเรียน" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ name: "โรงเรียนใหม่" }));
  });
});
