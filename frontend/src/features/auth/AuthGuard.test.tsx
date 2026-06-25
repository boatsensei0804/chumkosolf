import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { AuthGuard } from "./AuthGuard";

// --- mocks ---
const replaceMock = vi.fn();
let pathnameValue = "/";
let authValue: {
  isHydrated: boolean;
  isAuthenticated: boolean;
};

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock }),
  usePathname: () => pathnameValue,
}));

vi.mock("./AuthContext", () => ({
  useAuth: () => authValue,
}));

function renderGuard(): void {
  render(
    <AuthGuard>
      <div>เนื้อหาภายใน</div>
    </AuthGuard>,
  );
}

beforeEach(() => {
  replaceMock.mockClear();
  pathnameValue = "/";
  authValue = { isHydrated: true, isAuthenticated: false };
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("AuthGuard", () => {
  it("ยังไม่ hydrate → ไม่แสดงเนื้อหาและยังไม่ redirect", () => {
    authValue = { isHydrated: false, isAuthenticated: false };
    renderGuard();

    expect(screen.queryByText("เนื้อหาภายใน")).not.toBeInTheDocument();
    expect(replaceMock).not.toHaveBeenCalled();
  });

  it("ยังไม่ล็อกอิน + เข้าหน้าใน → เด้งไป /login และไม่แสดงเนื้อหา", () => {
    pathnameValue = "/";
    authValue = { isHydrated: true, isAuthenticated: false };
    renderGuard();

    expect(replaceMock).toHaveBeenCalledWith("/login");
    expect(screen.queryByText("เนื้อหาภายใน")).not.toBeInTheDocument();
  });

  it("ยังไม่ล็อกอิน + อยู่ /login → แสดงเนื้อหา (ฟอร์ม) ไม่ redirect", () => {
    pathnameValue = "/login";
    authValue = { isHydrated: true, isAuthenticated: false };
    renderGuard();

    expect(screen.getByText("เนื้อหาภายใน")).toBeInTheDocument();
    expect(replaceMock).not.toHaveBeenCalled();
  });

  it("ล็อกอินแล้ว + เข้าหน้าใน → แสดงเนื้อหา", () => {
    pathnameValue = "/";
    authValue = { isHydrated: true, isAuthenticated: true };
    renderGuard();

    expect(screen.getByText("เนื้อหาภายใน")).toBeInTheDocument();
    expect(replaceMock).not.toHaveBeenCalled();
  });

  it("ล็อกอินแล้ว + อยู่ /login → เด้งกลับหน้าแรกและไม่แสดงเนื้อหา", () => {
    pathnameValue = "/login";
    authValue = { isHydrated: true, isAuthenticated: true };
    renderGuard();

    expect(replaceMock).toHaveBeenCalledWith("/");
    expect(screen.queryByText("เนื้อหาภายใน")).not.toBeInTheDocument();
  });
});
