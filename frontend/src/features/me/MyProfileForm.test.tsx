import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render as rtlRender, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { MyProfileForm } from "./MyProfileForm";
import { type MyProfileFormValues } from "./formSchema";

// MyProfileForm ใช้ react-query (AddressCascade โหลดข้อมูลที่อยู่) → ต้องมี QueryClientProvider
function render(ui: ReactNode): ReturnType<typeof rtlRender> {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return rtlRender(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

function emptyValues(): MyProfileFormValues {
  return {
    prefix: "",
    firstName: "",
    lastName: "",
    birthDate: "",
    phone: "",
    email: "",
    address: { houseNo: "", moo: "", road: "", subdistrict: "", district: "", province: "", postalCode: "" },
  };
}

describe("MyProfileForm", () => {
  it("ไม่มีฟิลด์บัญชี/เลขบัตร (แก้เองไม่ได้)", () => {
    render(<MyProfileForm defaultValues={emptyValues()} onSubmit={vi.fn()} isSubmitting={false} />);
    expect(screen.queryByLabelText("ชื่อผู้ใช้")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("เลขบัตรประชาชน")).not.toBeInTheDocument();
    expect(screen.getByLabelText("ชื่อ")).toBeInTheDocument();
  });

  it("กรอกชื่อ/นามสกุลไม่ครบ → error ไทย ไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<MyProfileForm defaultValues={emptyValues()} onSubmit={onSubmit} isSubmitting={false} />);

    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลของฉัน" }));

    expect(await screen.findByText("กรุณากรอกชื่อ")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกนามสกุล")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit พร้อมค่า", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <MyProfileForm
        defaultValues={{ ...emptyValues(), firstName: "สมชาย", lastName: "ใจดี" }}
        onSubmit={onSubmit}
        isSubmitting={false}
      />,
    );

    await user.type(screen.getByLabelText("เบอร์โทร"), "0810000000");
    await user.click(screen.getByRole("button", { name: "บันทึกข้อมูลของฉัน" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ firstName: "สมชาย", lastName: "ใจดี", phone: "0810000000" }),
      expect.anything(),
    );
  });
});
