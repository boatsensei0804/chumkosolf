import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
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
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider>
        <App>
          <DashboardLayout>
            <div>เนื้อหาหน้า</div>
          </DashboardLayout>
        </App>
      </ConfigProvider>
    </QueryClientProvider>,
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
    expect(within(dialog).getByText("ตั้งค่าระบบ")).toBeInTheDocument();
  });

  it("ครูทั่วไป (ไม่ใช่ school admin) ไม่เห็นเมนูตั้งค่าระบบ", async () => {
    currentUser = makeUser({ role: "teacher", is_school_admin: false });
    const user = userEvent.setup();
    renderLayout();

    await user.click(screen.getByLabelText("เปิดเมนู"));
    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).queryByText("ตั้งค่าระบบ")).not.toBeInTheDocument();
  });

  it("\"ข้อมูลของฉัน\" อยู่ใต้ไอคอนบัญชี ไม่อยู่ใน sidebar", async () => {
    currentUser = makeUser({ role: "teacher", is_school_admin: false });
    const user = userEvent.setup();
    renderLayout();

    // ไม่อยู่ใน sidebar (drawer)
    await user.click(screen.getByLabelText("เปิดเมนู"));
    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).queryByText("ข้อมูลของฉัน")).not.toBeInTheDocument();
    expect(within(dialog).getByText("นักเรียนที่ปรึกษาของฉัน")).toBeInTheDocument();

    // อยู่ในเมนูไอคอนบัญชี
    await user.click(screen.getByLabelText("เมนูผู้ใช้"));
    expect(await screen.findByText("ข้อมูลของฉัน")).toBeInTheDocument();
  });

  it("กดออกจากระบบ → เรียก signOut", async () => {
    const user = userEvent.setup();
    renderLayout();

    await user.click(screen.getByLabelText("เมนูผู้ใช้"));
    await user.click(await screen.findByText("ออกจากระบบ"));

    expect(signOutMock).toHaveBeenCalledTimes(1);
  });
});
