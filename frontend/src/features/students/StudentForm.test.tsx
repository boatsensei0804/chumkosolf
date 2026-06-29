import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { emptyAddress } from "@/features/personnel/formSchema";

import { StudentForm } from "./StudentForm";
import type { CreateStudentFormValues } from "./formSchema";

function empty(): CreateStudentFormValues {
  return { nationalId: "", studentCode: "", status: "studying", classId: "", prefix: "", firstName: "", lastName: "", birthDate: "", phone: "", address: emptyAddress };
}

// StudentForm ดึงรายชื่อห้องผ่าน react-query → ต้องมี QueryClientProvider
function renderForm(node: ReactNode): void {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(<QueryClientProvider client={qc}>{node}</QueryClientProvider>);
}

describe("StudentForm", () => {
  it("render ฟิลด์หลัก", () => {
    renderForm(<StudentForm mode="create" defaultValues={empty()} onSubmit={vi.fn()} isSubmitting={false} submitLabel="เพิ่มนักเรียน" />);
    expect(screen.getByLabelText("รหัสนักเรียน")).toBeInTheDocument();
    expect(screen.getByLabelText("เลขบัตรประชาชน")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "เพิ่มนักเรียน" })).toBeInTheDocument();
  });

  it("กรอกไม่ครบ → error ไทย ไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    renderForm(<StudentForm mode="create" defaultValues={empty()} onSubmit={onSubmit} isSubmitting={false} submitLabel="เพิ่มนักเรียน" />);
    await user.click(screen.getByRole("button", { name: "เพิ่มนักเรียน" }));
    expect(await screen.findByText("กรุณากรอกรหัสนักเรียน")).toBeInTheDocument();
    expect(screen.getByText("เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    renderForm(<StudentForm mode="create" defaultValues={empty()} onSubmit={onSubmit} isSubmitting={false} submitLabel="เพิ่มนักเรียน" />);
    await user.type(screen.getByLabelText("รหัสนักเรียน"), "S001");
    await user.type(screen.getByLabelText("เลขบัตรประชาชน"), "1234567890123");
    await user.type(screen.getByLabelText("ชื่อ"), "กิตติ");
    await user.type(screen.getByLabelText("นามสกุล"), "เรียนดี");
    await user.click(screen.getByRole("button", { name: "เพิ่มนักเรียน" }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ studentCode: "S001", nationalId: "1234567890123", firstName: "กิตติ" }),
      expect.anything(),
    );
  });
});
