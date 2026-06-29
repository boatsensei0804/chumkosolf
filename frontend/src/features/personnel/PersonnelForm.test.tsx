import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render as rtlRender, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { PersonnelForm } from "./PersonnelForm";
import { emptyAddress, type CreatePersonnelFormValues } from "./formSchema";

// PersonnelForm ใช้ react-query (โหลดข้อมูลที่อยู่) → ต้องมี QueryClientProvider
function render(ui: ReactNode): ReturnType<typeof rtlRender> {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return rtlRender(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

function emptyValues(): CreatePersonnelFormValues {
  return {
    username: "",
    password: "",
    role: "teacher",
    nationalId: "",
    civilServantId: "",
    prefix: "",
    firstName: "",
    lastName: "",
    birthDate: "",
    phone: "",
    email: "",
    address: emptyAddress,
  };
}

describe("PersonnelForm (create)", () => {
  it("แสดงส่วนบัญชีผู้ใช้และข้อมูลส่วนตัว", () => {
    render(
      <PersonnelForm
        mode="create"
        defaultValues={emptyValues()}
        onSubmit={vi.fn()}
        isSubmitting={false}
        submitLabel="เพิ่มบุคลากร"
      />,
    );
    expect(screen.getByLabelText("ชื่อผู้ใช้")).toBeInTheDocument();
    expect(screen.getByLabelText("รหัสผ่าน")).toBeInTheDocument();
    expect(screen.getByLabelText("เลขบัตรประชาชน")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "เพิ่มบุคลากร" })).toBeInTheDocument();
  });

  it("กรอกไม่ครบ → แสดง error ไทย และไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <PersonnelForm
        mode="create"
        defaultValues={emptyValues()}
        onSubmit={onSubmit}
        isSubmitting={false}
        submitLabel="เพิ่มบุคลากร"
      />,
    );

    await user.click(screen.getByRole("button", { name: "เพิ่มบุคลากร" }));

    expect(await screen.findByText("เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกชื่อ")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit พร้อมค่า", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <PersonnelForm
        mode="create"
        defaultValues={emptyValues()}
        onSubmit={onSubmit}
        isSubmitting={false}
        submitLabel="เพิ่มบุคลากร"
      />,
    );

    await user.type(screen.getByLabelText("ชื่อผู้ใช้"), "kru.somchai");
    await user.type(screen.getByLabelText("รหัสผ่าน"), "password123");
    await user.type(screen.getByLabelText("เลขบัตรประชาชน"), "1234567890123");
    await user.type(screen.getByLabelText("ชื่อ"), "สมชาย");
    await user.type(screen.getByLabelText("นามสกุล"), "ใจดี");
    await user.click(screen.getByRole("button", { name: "เพิ่มบุคลากร" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        username: "kru.somchai",
        nationalId: "1234567890123",
        firstName: "สมชาย",
        lastName: "ใจดี",
      }),
      expect.anything(),
    );
  });
});

describe("PersonnelForm (edit)", () => {
  it("ไม่แสดงส่วนบัญชีผู้ใช้ และเลขบัตรไม่บังคับ", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <PersonnelForm
        mode="edit"
        defaultValues={{ ...emptyValues(), firstName: "สมหญิง", lastName: "ใจงาม" }}
        onSubmit={onSubmit}
        isSubmitting={false}
        submitLabel="บันทึกการแก้ไข"
      />,
    );

    expect(screen.queryByLabelText("ชื่อผู้ใช้")).not.toBeInTheDocument();

    // ไม่ได้กรอกเลขบัตร แต่ submit ได้ (edit ปล่อยว่าง = ไม่เปลี่ยน)
    await user.click(screen.getByRole("button", { name: "บันทึกการแก้ไข" }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
  });
});
