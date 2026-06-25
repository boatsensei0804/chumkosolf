"use client";

import {
  ApartmentOutlined,
  AppstoreOutlined,
  ArrowRightOutlined,
  DollarOutlined,
  IdcardOutlined,
  SafetyOutlined,
  ScheduleOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { Tag } from "antd";
import { useRouter } from "next/navigation";
import type { ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { usePersonnelList } from "@/features/personnel/hooks";
import { menuItemsForUser } from "@/features/navigation/menu";
import { userRoleSchema, workGroupLabel, type UserRole } from "@/shared/schemas/enums";
import { ACCENTS, moduleAccent, type Accent } from "@/shared/ui/accent";
import { SectionCard } from "@/shared/ui/SectionCard";

const QUICK: Record<string, { icon: ReactNode; desc: string }> = {
  personnel: { icon: <TeamOutlined />, desc: "ครูและบุคลากร" },
  students: { icon: <UserOutlined />, desc: "นักเรียน/ผู้ปกครอง" },
  attendance: { icon: <ScheduleOutlined />, desc: "เช็คชื่อ/ความประพฤติ" },
  budget: { icon: <DollarOutlined />, desc: "งบประมาณ/โครงการ" },
  settings: { icon: <SettingOutlined />, desc: "ตั้งค่าระบบ" },
};

const ROLE_LABEL: Record<UserRole, string> = {
  super_admin: "ผู้ดูแลระบบสูงสุด",
  teacher: "ครู",
  executive: "ผู้บริหาร",
  student: "นักเรียน",
};

// KPI card — ตัวเลขนำ + แถบสี accent ด้านบน (data-dense, เด่นชัด)
function Kpi({
  icon,
  label,
  value,
  hint,
  accent,
}: {
  icon: ReactNode;
  label: string;
  value: ReactNode;
  hint?: string;
  accent: Accent;
}): ReactNode {
  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div className={`h-1 ${ACCENTS[accent].strip}`} />
      <div className="flex items-start justify-between p-3.5">
        <div className="min-w-0">
          <div className="text-xs text-slate-500">{label}</div>
          <div className="mt-1 text-2xl font-bold leading-none text-brand-navy">{value}</div>
          {hint && <div className="mt-1.5 truncate text-xs text-slate-400">{hint}</div>}
        </div>
        <span className={`flex h-9 w-9 items-center justify-center rounded-lg ${ACCENTS[accent].chip}`}>
          {icon}
        </span>
      </div>
    </div>
  );
}

export default function DashboardHomePage(): ReactNode {
  const { user } = useAuth();
  const router = useRouter();

  const destinations = user ? menuItemsForUser(user).filter((i) => i.key !== "home") : [];
  const canPersonnel = destinations.some((i) => i.key === "personnel");
  const { data: personnelData } = usePersonnelList(1, 5, canPersonnel);

  if (!user) return null;

  const role = userRoleSchema.catch("student").parse(user.role);
  const recent = personnelData?.items ?? [];

  return (
    <div className="flex flex-col gap-4">
      {/* welcome บรรทัดเดียว (title/date อยู่บน topbar แล้ว) */}
      <div>
        <h2 className="text-lg font-bold text-slate-800">สวัสดี, {user.username}</h2>
        <p className="text-sm text-slate-500">
          {ROLE_LABEL[role]} · โรงเรียนชุมโคพิทยาคม
        </p>
      </div>

      {/* KPI row */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {canPersonnel && (
          <Kpi
            accent="blue"
            icon={<TeamOutlined />}
            label="บุคลากรในระบบ"
            value={<span className="num">{personnelData?.total ?? "—"}</span>}
            hint="ครูและผู้บริหาร"
          />
        )}
        <Kpi
          accent="violet"
          icon={<ApartmentOutlined />}
          label="กลุ่มงานที่สังกัด"
          value={<span className="num">{user.work_groups.length}</span>}
          hint={user.is_school_admin ? "ผู้ดูแล (เห็นทุกกลุ่ม)" : undefined}
        />
        <Kpi
          accent="emerald"
          icon={<IdcardOutlined />}
          label="บทบาท"
          value={<span className="text-base font-semibold">{ROLE_LABEL[role]}</span>}
        />
        <Kpi
          accent="amber"
          icon={<SafetyOutlined />}
          label="สิทธิ์"
          value={<span className="text-base font-semibold">{user.is_school_admin ? "ผู้ดูแลระบบ" : "สมาชิก"}</span>}
        />
      </div>

      {/* main grid */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <SectionCard icon={<AppstoreOutlined />} title="เมนูการทำงาน" description="เข้าถึงงานตามสิทธิ์ของคุณ">
            <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
              {destinations.map((item) => {
                const meta = QUICK[item.key];
                const icon = meta?.icon ?? <AppstoreOutlined />;
                const acc = moduleAccent(item.key);
                if (!item.available) {
                  return (
                    <div
                      key={item.key}
                      className="flex items-center gap-2.5 rounded-lg border border-slate-200 bg-slate-50/60 px-3 py-2 opacity-70"
                    >
                      <span className="flex h-8 w-8 items-center justify-center rounded-md bg-slate-100 text-slate-400">
                        {icon}
                      </span>
                      <span className="flex-1 text-sm font-medium text-slate-500">{item.label}</span>
                      <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] text-slate-400">
                        เร็ว ๆ นี้
                      </span>
                    </div>
                  );
                }
                return (
                  <button
                    key={item.key}
                    type="button"
                    onClick={() => router.push(item.path)}
                    className="group flex items-center gap-2.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-left transition-colors hover:border-slate-300 hover:bg-slate-50"
                  >
                    <span className={`flex h-8 w-8 items-center justify-center rounded-md ${ACCENTS[acc].chip}`}>
                      {icon}
                    </span>
                    <span className="flex-1">
                      <span className="block text-sm font-semibold text-slate-800">{item.label}</span>
                      <span className="block text-xs text-slate-500">{meta?.desc}</span>
                    </span>
                    <ArrowRightOutlined className="text-slate-300 transition-all group-hover:translate-x-0.5 group-hover:text-brand" />
                  </button>
                );
              })}
            </div>
          </SectionCard>
        </div>

        <div className="flex flex-col gap-4">
          {canPersonnel && (
            <SectionCard
              icon={<IdcardOutlined />}
              title="บุคลากรล่าสุด"
              extra={
                <button type="button" onClick={() => router.push("/personnel")} className="text-xs text-brand">
                  ดูทั้งหมด
                </button>
              }
            >
              {recent.length === 0 ? (
                <p className="py-1 text-sm text-slate-400">ยังไม่มีข้อมูลบุคลากร</p>
              ) : (
                <ul className="-my-1 divide-y divide-slate-100">
                  {recent.map((p) => (
                    <li key={p.id}>
                      <button
                        type="button"
                        onClick={() => router.push(`/personnel/${p.id}/edit`)}
                        className="flex w-full items-center gap-2.5 py-2 text-left transition-colors hover:bg-slate-50"
                      >
                        <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand/10 text-[11px] font-semibold text-brand">
                          {(p.first_name.charAt(0) || "?").toUpperCase()}
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="block truncate text-sm text-slate-700">
                            {p.prefix}
                            {p.first_name} {p.last_name}
                          </span>
                          <span className="block text-[11px] text-slate-400">@{p.username}</span>
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </SectionCard>
          )}

          <SectionCard icon={<ApartmentOutlined />} title="กลุ่มงานที่สังกัด" accent="emerald">
            {user.work_groups.length > 0 ? (
              <div className="flex flex-wrap gap-1.5">
                {user.work_groups.map((g) => (
                  <Tag key={g.work_group_id} color="blue" bordered={false}>
                    {workGroupLabel[g.code]}
                    {g.is_group_admin && " · หัวหน้า"}
                  </Tag>
                ))}
              </div>
            ) : (
              <p className="text-sm text-slate-400">ยังไม่ได้สังกัดกลุ่มงาน — โปรดให้ผู้ดูแลมอบหมาย</p>
            )}
          </SectionCard>
        </div>
      </div>
    </div>
  );
}
