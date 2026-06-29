import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render as rtlRender, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { AdviseeEditForm } from "./AdviseeEditForm";
import { type AdviseeFormValues } from "./adviseeForm";

function render(ui: ReactNode): ReturnType<typeof rtlRender> {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return rtlRender(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

function values(overrides: Partial<AdviseeFormValues> = {}): AdviseeFormValues {
  return {
    prefix: "",
    firstName: "",
    lastName: "",
    birthDate: "",
    phone: "",
    address: { houseNo: "", moo: "", road: "", subdistrict: "", district: "", province: "", postalCode: "" },
    ...overrides,
  };
}

describe("AdviseeEditForm", () => {
  it("ไม่มีฟิลด์ทะเบียน (รหัสนักเรียน/สถานะ/เลขบัตร)", () => {
    render(<AdviseeEditForm defaultValues={values()} onSubmit={vi.fn()} isSubmitting={false} />);
    expect(screen.queryByLabelText("รหัสนักเรียน")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("เลขบัตรประชาชน")).not.toBeInTheDocument();
    expect(screen.getByLabelText("ชื่อ")).toBeInTheDocument();
  });

  it("กรอกชื่อ/นามสกุลไม่ครบ → error ไทย ไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AdviseeEditForm defaultValues={values()} onSubmit={onSubmit} isSubmitting={false} />);

    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลนักเรียน" }));

    expect(await screen.findByText("กรุณากรอกชื่อ")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกนามสกุล")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit พร้อมค่า", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <AdviseeEditForm
        defaultValues={values({ firstName: "เด็กชาย", lastName: "ใจดี" })}
        onSubmit={onSubmit}
        isSubmitting={false}
      />,
    );

    await user.type(screen.getByLabelText("เบอร์โทร"), "0810000000");
    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลนักเรียน" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ firstName: "เด็กชาย", lastName: "ใจดี", phone: "0810000000" }),
      expect.anything(),
    );
  });
});
