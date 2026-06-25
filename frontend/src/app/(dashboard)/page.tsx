"use client";

import {
  AppstoreOutlined,
  ArrowRightOutlined,
  DollarOutlined,
  ScheduleOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { useRouter } from "next/navigation";
import type { ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { menuItemsForUser } from "@/features/navigation/menu";
import { workGroupLabel } from "@/shared/schemas/enums";

// ไอคอน + คำอธิบายของแต่ละปลายทาง (ใช้บนการ์ดทางลัด)
const QUICK: Record<string, { icon: ReactNode; desc: string }> = {
  personnel: { icon: <TeamOutlined />, desc: "จัดการข้อมูลครูและบุคลากร" },
  students: { icon: <UserOutlined />, desc: "ข้อมูลนักเรียนและผู้ปกครอง" },
  attendance: { icon: <ScheduleOutlined />, desc: "เช็คชื่อและคะแนนความประพฤติ" },
  budget: { icon: <DollarOutlined />, desc: "งบประมาณและโครงการ" },
  settings: { icon: <SettingOutlined />, desc: "ตั้งค่าระบบและสิทธิ์ผู้ใช้" },
};

export default function DashboardHomePage(): ReactNode {
  const { user } = useAuth();
  const router = useRouter();

  if (!user) return null;

  const today = new Intl.DateTimeFormat("th-TH", { dateStyle: "full" }).format(new Date());
  // การ์ดทางลัด = เมนูที่ผู้ใช้มีสิทธิ์ ยกเว้นหน้าแรก
  const destinations = menuItemsForUser(user).filter((i) => i.key !== "home");

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6">
      {/* hero ทักทาย — โมเมนต์เด่นชิ้นเดียวของหน้า */}
      <section className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-brand to-brand-deep px-6 py-7 text-white shadow-sm md:px-8">
        <div className="pointer-events-none absolute -right-10 -top-10 h-44 w-44 rounded-full bg-white/10" />
        <div className="pointer-events-none absolute -bottom-16 right-16 h-40 w-40 rounded-full bg-brand-cyan/20" />
        <div className="relative">
          <p className="text-sm text-white/70">{today}</p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight md:text-3xl">
            สวัสดี, {user.username}
          </h1>
          <p className="mt-1 text-sm text-white/80">
            ยินดีต้อนรับสู่ระบบบริหารจัดการโรงเรียนชุมโค
          </p>
          {user.work_groups.length > 0 && (
            <div className="mt-4 flex flex-wrap gap-2">
              {user.work_groups.map((g) => (
                <span
                  key={g.work_group_id}
                  className="rounded-full bg-white/15 px-3 py-1 text-xs font-medium text-white ring-1 ring-inset ring-white/20"
                >
                  {workGroupLabel[g.code]}
                  {g.is_group_admin && " · หัวหน้ากลุ่ม"}
                </span>
              ))}
            </div>
          )}
        </div>
      </section>

      {/* ทางลัดไปยังงานที่มีสิทธิ์ */}
      <section>
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-400">
          เมนูการทำงาน
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {destinations.map((item) => {
            const meta = QUICK[item.key];
            const icon = meta?.icon ?? <AppstoreOutlined />;
            const desc = meta?.desc ?? "";

            if (!item.available) {
              return (
                <div
                  key={item.key}
                  className="flex items-start gap-3 rounded-xl border border-slate-200 bg-white p-4 opacity-70"
                >
                  <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-slate-100 text-lg text-slate-400">
                    {icon}
                  </span>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-slate-500">{item.label}</span>
                      <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-medium text-slate-400">
                        เร็ว ๆ นี้
                      </span>
                    </div>
                    <p className="mt-0.5 text-xs text-slate-400">{desc}</p>
                  </div>
                </div>
              );
            }

            return (
              <button
                key={item.key}
                type="button"
                onClick={() => router.push(item.path)}
                className="group flex items-start gap-3 rounded-xl border border-slate-200 bg-white p-4 text-left transition-all hover:-translate-y-0.5 hover:border-brand/30 hover:shadow-md"
              >
                <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand/10 text-lg text-brand transition-colors group-hover:bg-brand group-hover:text-white">
                  {icon}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-slate-800">{item.label}</span>
                    <ArrowRightOutlined className="text-slate-300 transition-all group-hover:translate-x-0.5 group-hover:text-brand" />
                  </div>
                  <p className="mt-0.5 text-xs text-slate-500">{desc}</p>
                </div>
              </button>
            );
          })}
        </div>
      </section>

      {user.work_groups.length === 0 && (
        <section className="rounded-xl border border-amber-200 bg-amber-50 p-4">
          <p className="text-sm text-amber-700">
            คุณยังไม่ได้สังกัดกลุ่มงาน — โปรดให้ผู้ดูแลระบบมอบหมายกลุ่มงาน
            เพื่อเปิดสิทธิ์การเข้าถึงเมนูต่าง ๆ
          </p>
        </section>
      )}

      <p className="text-center text-xs text-slate-400">
        ระบบอยู่ระหว่างพัฒนา (Phase 1) · เมนูที่ยังไม่เปิดจะทยอยพร้อมใช้งานในเฟสถัดไป
      </p>
    </div>
  );
}
