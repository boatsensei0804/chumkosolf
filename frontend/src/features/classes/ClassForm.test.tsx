import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ClassForm } from "./ClassForm";

describe("ClassForm", () => {
  it("กรอกไม่ครบ → error ไทย ไม่เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<ClassForm defaultValues={{ gradeLevel: "", roomName: "" }} onSubmit={onSubmit} isSubmitting={false} submitLabel="เพิ่มห้องเรียน" />);
    await user.click(screen.getByRole("button", { name: "เพิ่มห้องเรียน" }));
    expect(await screen.findByText("กรุณากรอกระดับชั้น")).toBeInTheDocument();
    expect(screen.getByText("กรุณากรอกห้อง")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("กรอกครบ → เรียก onSubmit", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<ClassForm defaultValues={{ gradeLevel: "", roomName: "" }} onSubmit={onSubmit} isSubmitting={false} submitLabel="เพิ่มห้องเรียน" />);
    await user.type(screen.getByLabelText("ระดับชั้น"), "ม.1");
    await user.type(screen.getByLabelText("ห้อง"), "1/1");
    await user.click(screen.getByRole("button", { name: "เพิ่มห้องเรียน" }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ gradeLevel: "ม.1", roomName: "1/1" }), expect.anything());
  });
});
