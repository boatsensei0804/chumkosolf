import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App, ConfigProvider } from "antd";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { UserInfo } from "@/shared/schemas/auth";

import { DashboardLayout } from "./DashboardLayout";

const signOutMock = vi.fn();
const pushMock = vi.fn();
let currentUser: UserInfo | null;

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn() }),
  usePathname: () => "/",
}));

vi.mock("@/features/auth/AuthContext", () => ({
  useAuth: () => ({ user: currentUser, signOut: signOutMock }),
}));

function makeUser(overrides: Partial<UserInfo> = {}): UserInfo {
  return {
    id: "u1",
    username: "superadmin",
    role: "super_admin",
    school_id: "s1",
    is_school_admin: true,
    work_groups: [],
    ...overrides,
  };
}

function renderLayout(): void {
  render(
    <ConfigProvider>
      <App>
        <DashboardLayout>
          <div>เนื้อหาหน้า</div>
        </DashboardLayout>
      </App>
    </ConfigProvider>,
  );
}

beforeEach(() => {
  signOutMock.mockClear();
  pushMock.mockClear();
  currentUser = makeUser();
});

describe("DashboardLayout", () => {
  it("แสดงชื่อผู้ใช้และเนื้อหา children", () => {
    renderLayout();
    expect(screen.getByText("superadmin")).toBeInTheDocument();
    expect(screen.getByText("เนื้อหาหน้า")).toBeInTheDocument();
  });

  it("เปิดเมนู → super_admin เห็นทั้งหน้าแรกและตั้งค่าระบบ", async () => {
    const user = userEvent.setup();
    renderLayout();

    await user.click(screen.getByLabelText("เปิดเมนู"));
    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("หน้าแรก")).toBeInTheDocument();
    // เมนูที่ยังไม่เปิด: label + ป้าย "เร็ว ๆ นี้" แยกกัน
    expect(within(dialog).getByText("ตั้งค่าระบบ")).toBeInTheDocument();
    expect(within(dialog).getAllByText("เร็ว ๆ นี้").length).toBeGreaterThan(0);
  });

  it("ครูทั่วไป (ไม่ใช่ school admin) ไม่เห็นเมนูตั้งค่าระบบ", async () => {
    currentUser = makeUser({ role: "teacher", is_school_admin: false });
    const user = userEvent.setup();
    renderLayout();

    await user.click(screen.getByLabelText("เปิดเมนู"));
    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).queryByText("ตั้งค่าระบบ")).not.toBeInTheDocument();
  });

  it("กดออกจากระบบ → เรียก signOut", async () => {
    const user = userEvent.setup();
    renderLayout();

    await user.click(screen.getByLabelText("เมนูผู้ใช้"));
    await user.click(await screen.findByText("ออกจากระบบ"));

    expect(signOutMock).toHaveBeenCalledTimes(1);
  });
});
