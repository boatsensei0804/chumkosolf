import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "antd";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import KioskPage from "./page";

const reindexMutate = vi.fn();
const recognizeMutate = vi.fn();

vi.mock("@/features/face/hooks", () => ({
  useReindexFace: () => ({ mutate: reindexMutate, isPending: false }),
  useRecognizeFace: () => ({ mutate: recognizeMutate, isPending: false }),
}));

// ครู (กลุ่มวิชาการ ไม่ใช่ admin) → เห็นปุ่มอัปเดตฐานใบหน้า, ไม่เห็นแท็บบัญชี
vi.mock("@/features/auth/AuthContext", () => ({
  useAuth: () => ({ user: { role: "teacher", is_school_admin: false, work_groups: [{ code: "academic" }] } }),
}));

beforeEach(() => {
  reindexMutate.mockClear();
  recognizeMutate.mockClear();
  // mock กล้อง
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: vi.fn().mockResolvedValue({ getTracks: () => [] }),
      enumerateDevices: vi.fn().mockResolvedValue([]),
    },
  });
});

function renderPage(ui: ReactNode): ReturnType<typeof render> {
  return render(<App>{ui}</App>);
}

describe("KioskPage", () => {
  it("แสดงหัวข้อ + ปุ่มสแกน + ปุ่มอัปเดตฐานใบหน้า", () => {
    renderPage(<KioskPage />);
    expect(screen.getByText("สแกนหน้าเข้าเรียน")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /สแกน/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /อัปเดตฐานใบหน้า/ })).toBeInTheDocument();
  });

  it("กดอัปเดตฐานใบหน้า → เรียก reindex", async () => {
    const user = userEvent.setup();
    renderPage(<KioskPage />);
    await user.click(screen.getByRole("button", { name: /อัปเดตฐานใบหน้า/ }));
    expect(reindexMutate).toHaveBeenCalled();
  });
});
