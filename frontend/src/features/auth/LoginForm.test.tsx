import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LoginForm } from "./LoginForm";

describe("LoginForm", () => {
  it("render ฟิลด์ครบและ default รหัสโรงเรียนเป็น CHUMKO", () => {
    render(<LoginForm onSubmit={vi.fn()} isSubmitting={false} />);

    expect(screen.getByLabelText("รหัสโรงเรียน")).toHaveValue("CHUMKO");
    expect(screen.getByLabelText("ชื่อผู้ใช้")).toBeInTheDocument();
    expect(screen.getByLabelText("รหัสผ่าน")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "เข้าสู่ระบบ" })).toBeInTheDocument();
  });

  it("กรอกไม่ครบ → แสดง error ภาษาไทย และไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} isSubmitting={false} />);

    // ลบ username/password ให้ว่าง แล้วกดเข้าสู่ระบบ
    await user.click(screen.getByRole("button", { name: "เข้าสู่ระบบ" }));

    expect(await screen.findByText("กรุณากรอกชื่อผู้ใช้")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกรหัสผ่าน")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit พร้อมค่าที่กรอก", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} isSubmitting={false} />);

    await user.type(screen.getByLabelText("ชื่อผู้ใช้"), "superadmin");
    await user.type(screen.getByLabelText("รหัสผ่าน"), "admin1234");
    await user.click(screen.getByRole("button", { name: "เข้าสู่ระบบ" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      { schoolCode: "CHUMKO", username: "superadmin", password: "admin1234" },
      expect.anything(),
    );
  });

  it("มี errorMessage → แสดง alert", () => {
    render(
      <LoginForm
        onSubmit={vi.fn()}
        isSubmitting={false}
        errorMessage="ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent(
      "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง",
    );
  });
});
