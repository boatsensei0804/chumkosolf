"use client";

import {
  AppstoreOutlined,
  DollarOutlined,
  HomeOutlined,
  LogoutOutlined,
  MenuOutlined,
  ReadOutlined,
  ScheduleOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { App, Drawer, Dropdown } from "antd";
import type { MenuProps } from "antd";
import { usePathname, useRouter } from "next/navigation";
import { useMemo, useState, type ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { userRoleSchema, type UserRole } from "@/shared/schemas/enums";

import { menuItemsForUser, type MenuItemConfig } from "./menu";

const MENU_ICONS: Record<string, ReactNode> = {
  home: <HomeOutlined />,
  personnel: <TeamOutlined />,
  students: <UserOutlined />,
  attendance: <ScheduleOutlined />,
  budget: <DollarOutlined />,
  settings: <SettingOutlined />,
};

const ROLE_LABEL: Record<UserRole, string> = {
  super_admin: "ผู้ดูแลระบบสูงสุด",
  teacher: "ครู",
  executive: "ผู้บริหาร",
  student: "นักเรียน",
};

// แบรนด์มาร์ก: โลโก้ gradient ฟ้า + wordmark ไทย (signature ของระบบ)
function BrandMark(): ReactNode {
  return (
    <div className="flex items-center gap-2.5">
      <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand to-brand-deep text-white shadow-sm">
        <ReadOutlined className="text-lg" />
      </div>
      <div className="leading-tight">
        <div className="text-[15px] font-bold tracking-tight text-slate-800">ชุมโค</div>
        <div className="text-[11px] text-slate-400">ระบบบริหารโรงเรียน</div>
      </div>
    </div>
  );
}

// รายการเมนู (custom เพื่อคุม active state / coming-soon ได้เต็มที่)
function NavList({
  items,
  activeKey,
  onNavigate,
}: {
  items: MenuItemConfig[];
  activeKey: string;
  onNavigate: (item: MenuItemConfig) => void;
}): ReactNode {
  return (
    <nav className="flex flex-col gap-1 px-3 py-4">
      {items.map((item) => {
        const icon = MENU_ICONS[item.key] ?? <AppstoreOutlined />;
        const active = item.key === activeKey;

        if (!item.available) {
          return (
            <div
              key={item.key}
              className="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-2.5 text-slate-400"
            >
              <span className="text-base">{icon}</span>
              <span className="flex-1 text-sm">{item.label}</span>
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-medium text-slate-400">
                เร็ว ๆ นี้
              </span>
            </div>
          );
        }

        return (
          <button
            key={item.key}
            type="button"
            onClick={() => onNavigate(item)}
            aria-current={active ? "page" : undefined}
            className={[
              "group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors",
              active
                ? "bg-brand/10 font-semibold text-brand-deep"
                : "text-slate-600 hover:bg-slate-50 hover:text-slate-900",
            ].join(" ")}
          >
            {active && (
              <span className="absolute left-0 top-1/2 h-5 w-1 -translate-y-1/2 rounded-r-full bg-brand" />
            )}
            <span className={["text-base", active ? "text-brand" : "text-slate-400 group-hover:text-slate-600"].join(" ")}>
              {icon}
            </span>
            <span className="flex-1 text-left">{item.label}</span>
          </button>
        );
      })}
    </nav>
  );
}

export function DashboardLayout({ children }: { children: ReactNode }): ReactNode {
  const router = useRouter();
  const pathname = usePathname();
  const { message } = App.useApp();
  const { user, signOut } = useAuth();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const items = useMemo(() => (user ? menuItemsForUser(user) : []), [user]);

  const activeKey = useMemo(() => {
    const match = items.find(
      (i) => i.path === pathname || (i.path !== "/" && pathname.startsWith(i.path)),
    );
    return match?.key ?? "home";
  }, [items, pathname]);

  const pageTitle = items.find((i) => i.key === activeKey)?.label ?? "หน้าแรก";

  const handleNavigate = (item: MenuItemConfig): void => {
    router.push(item.path);
    setDrawerOpen(false);
  };

  const handleSignOut = (): void => {
    signOut();
    message.success("ออกจากระบบแล้ว");
  };

  const userMenu: MenuProps["items"] = [
    { key: "logout", icon: <LogoutOutlined />, label: "ออกจากระบบ", danger: true, onClick: handleSignOut },
  ];

  const role = user ? userRoleSchema.catch("student").parse(user.role) : "student";
  const roleLabel = user ? ROLE_LABEL[role] : "";
  const initial = user?.username.charAt(0).toUpperCase() ?? "?";

  const today = new Intl.DateTimeFormat("th-TH", { dateStyle: "full" }).format(new Date());

  const nav = <NavList items={items} activeKey={activeKey} onNavigate={handleNavigate} />;

  return (
    <div className="flex min-h-dvh bg-slate-50">
      {/* sidebar (desktop) */}
      <aside className="sticky top-0 hidden h-dvh w-64 shrink-0 flex-col border-r border-slate-200 bg-white lg:flex">
        <div className="flex h-16 items-center border-b border-slate-100 px-5">
          <BrandMark />
        </div>
        <div className="flex-1 overflow-y-auto">{nav}</div>
      </aside>

      {/* drawer (mobile) */}
      <Drawer
        placement="left"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={280}
        styles={{ body: { padding: 0 }, header: { display: "none" } }}
      >
        <div className="flex h-16 items-center border-b border-slate-100 px-5">
          <BrandMark />
        </div>
        {nav}
      </Drawer>

      {/* main column */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-slate-200 bg-white/90 px-4 backdrop-blur md:px-6">
          <div className="flex items-center gap-3">
            <button
              type="button"
              aria-label="เปิดเมนู"
              onClick={() => setDrawerOpen(true)}
              className="flex h-9 w-9 items-center justify-center rounded-lg text-slate-600 hover:bg-slate-100 lg:hidden"
            >
              <MenuOutlined />
            </button>
            <div className="leading-tight">
              <h1 className="text-base font-semibold text-slate-800">{pageTitle}</h1>
              <p className="hidden text-xs text-slate-400 sm:block">{today}</p>
            </div>
          </div>

          <Dropdown menu={{ items: userMenu }} trigger={["click"]} placement="bottomRight">
            <button
              type="button"
              aria-label="เมนูผู้ใช้"
              className="flex items-center gap-2.5 rounded-xl py-1 pl-1 pr-2 transition-colors hover:bg-slate-100"
            >
              <span className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-brand to-brand-deep text-sm font-semibold text-white">
                {initial}
              </span>
              <span className="hidden text-left leading-tight sm:block">
                <span className="block text-sm font-medium text-slate-800">{user?.username}</span>
                <span className="block text-xs text-slate-400">{roleLabel}</span>
              </span>
            </button>
          </Dropdown>
        </header>

        <main className="flex-1 p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}
