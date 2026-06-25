"use client";

import {
  AppstoreOutlined,
  DollarOutlined,
  DownOutlined,
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

// สีไอคอนเมนูต่อหมวด (บนพื้น navy ใช้เฉดสว่าง) ให้เด่น/จำง่าย
const NAV_ICON_COLOR: Record<string, string> = {
  home: "text-sky-400",
  personnel: "text-sky-400",
  students: "text-emerald-400",
  attendance: "text-violet-400",
  budget: "text-amber-400",
  settings: "text-slate-400",
};

const ROLE_LABEL: Record<UserRole, string> = {
  super_admin: "ผู้ดูแลระบบสูงสุด",
  teacher: "ครู",
  executive: "ผู้บริหาร",
  student: "นักเรียน",
};

// แบรนด์มาร์กบนพื้น navy (sidebar)
function BrandMark(): ReactNode {
  return (
    <div className="flex items-center gap-2.5">
      <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand text-white">
        <ReadOutlined />
      </div>
      <div className="leading-tight">
        <div className="text-sm font-bold tracking-tight text-white">ชุมโค</div>
        <div className="text-[11px] text-slate-400">ระบบบริหารโรงเรียน</div>
      </div>
    </div>
  );
}

// เมนู (dark sidebar) — active = แถบ accent + พื้นสว่างจาง + ตัวขาว
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
    <nav className="flex flex-col gap-0.5 px-2.5 py-3">
      <div className="px-2 pb-1.5 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
        เมนู
      </div>
      {items.map((item) => {
        const icon = MENU_ICONS[item.key] ?? <AppstoreOutlined />;
        const active = item.key === activeKey;

        if (!item.available) {
          return (
            <div
              key={item.key}
              className="flex cursor-not-allowed items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-slate-500"
            >
              <span className="text-base">{icon}</span>
              <span className="flex-1">{item.label}</span>
              <span className="rounded bg-white/5 px-1.5 py-0.5 text-[10px] text-slate-500">
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
              "flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors",
              active
                ? "bg-brand font-medium text-white shadow-sm"
                : "text-slate-300 hover:bg-white/5 hover:text-white",
            ].join(" ")}
          >
            <span className={active ? "text-white" : (NAV_ICON_COLOR[item.key] ?? "text-slate-400")}>
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
  const today = new Intl.DateTimeFormat("th-TH", { dateStyle: "long" }).format(new Date());

  const sidebarInner = (
    <div className="flex h-full flex-col bg-brand-navy">
      <div className="flex h-14 items-center border-b border-white/10 px-4">
        <BrandMark />
      </div>
      <div className="flex-1 overflow-y-auto">
        <NavList items={items} activeKey={activeKey} onNavigate={handleNavigate} />
      </div>
    </div>
  );

  return (
    <div className="flex min-h-dvh bg-slate-50">
      {/* sidebar (desktop) — navy enterprise */}
      <aside className="sticky top-0 hidden h-dvh w-60 shrink-0 lg:block">{sidebarInner}</aside>

      {/* drawer (mobile) */}
      <Drawer
        placement="left"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={240}
        styles={{ body: { padding: 0 }, header: { display: "none" }, content: { background: "#0F172A" } }}
      >
        {sidebarInner}
      </Drawer>

      {/* main column */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4">
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              aria-label="เปิดเมนู"
              onClick={() => setDrawerOpen(true)}
              className="flex h-9 w-9 items-center justify-center rounded-md text-slate-600 hover:bg-slate-100 lg:hidden"
            >
              <MenuOutlined />
            </button>
            <h1 className="text-sm font-semibold text-slate-800">{pageTitle}</h1>
            <span className="num hidden text-xs text-slate-400 sm:inline">· {today}</span>
          </div>

          <Dropdown menu={{ items: userMenu }} trigger={["click"]} placement="bottomRight">
            <button
              type="button"
              aria-label="เมนูผู้ใช้"
              className="flex items-center gap-2 rounded-md px-1.5 py-1 transition-colors hover:bg-slate-100"
            >
              <span className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-navy text-xs font-semibold text-white">
                {initial}
              </span>
              <span className="hidden text-left leading-tight sm:block">
                <span className="block text-sm font-medium text-slate-800">{user?.username}</span>
                <span className="block text-[11px] text-slate-400">{roleLabel}</span>
              </span>
              <DownOutlined className="hidden text-[10px] text-slate-400 sm:inline" />
            </button>
          </Dropdown>
        </header>

        <main className="flex-1 p-4">{children}</main>
      </div>
    </div>
  );
}
