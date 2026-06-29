"use client";

import {
  ApartmentOutlined,
  AppstoreOutlined,
  ArrowRightOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  IdcardOutlined,
  SafetyOutlined,
  ScanOutlined,
  ScheduleOutlined,
  SettingOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { Tag } from "antd";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { HomeTeacherSummary } from "@/features/dashboard/HomeTeacherSummary";
import { useDashboard } from "@/features/dashboard/hooks";
import { usePersonnelList } from "@/features/personnel/hooks";
import { menuItemsForUser } from "@/features/navigation/menu";
import { userRoleSchema, workGroupLabel, type UserRole } from "@/shared/schemas/enums";
import { ACCENTS, moduleAccent, type Accent } from "@/shared/ui/accent";
import { SectionCard } from "@/shared/ui/SectionCard";

const QUICK: Record<string, { icon: ReactNode; desc: string }> = {
  my_advisees: { icon: <TeamOutlined />, desc: "นักเรียนที่ปรึกษาของฉัน" },
  people: { icon: <IdcardOutlined />, desc: "บุคลากร/นักเรียน/ผู้ปกครอง" },
  classes: { icon: <ApartmentOutlined />, desc: "ห้องเรียน/ค้นหานักเรียน" },
  kiosk: { icon: <ScanOutlined />, desc: "สแกนหน้าเข้าเรียน" },
  attendance: { icon: <ScheduleOutlined />, desc: "เช็คชื่อ/ความประพฤติ" },
  budget: { icon: <DollarOutlined />, desc: "งบประมาณ/โครงการ" },
  settings: { icon: <SettingOutlined />, desc: "ตั้งค่าระบบ" },
};

const ROLE_LABEL: Record<UserRole, string> = {
  super_admin: "ผู้ดูแลระบบสูงสุด",
  teacher: "ครู",
  executive: "ผู้บริหาร",
  student: "นักเรียน",
  kiosk: "เครื่องสแกนหน้า",
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

  const destinations = user
    ? menuItemsForUser(user).filter((i) => i.key !== "home" && !i.inAccountMenu)
    : [];
  const canPersonnel = destinations.some((i) => i.key === "personnel");
  const { data: personnelData } = usePersonnelList(1, 5, "", canPersonnel);
  const { data: dashboard } = useDashboard();

  // บัญชี kiosk ใช้ได้แค่หน้าสแกน → เด้งไปที่นั่น
  useEffect(() => {
    if (user?.role === "kiosk") router.replace("/kiosk");
  }, [user, router]);

  if (!user) return null;

  const role = userRoleSchema.catch("student").parse(user.role);
  const recent = personnelData?.items ?? [];

  // KPI งานครู — คำนวณจากข้อมูล dashboard (ที่ปรึกษา + ตารางสอน)
  const todayLessons = dashboard
    ? dashboard.slots.filter((s) => s.day_of_week === dashboard.today_weekday).length
    : 0;
  const weekLessons = dashboard?.slots.length ?? 0;
  const hasTimetable = weekLessons > 0;
  const isAdvisor = dashboard?.is_advisor ?? false;

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
        {isAdvisor && (
          <Kpi
            accent="emerald"
            icon={<TeamOutlined />}
            label="นักเรียนที่ปรึกษา"
            value={<span className="num">{dashboard?.advisee_count ?? 0}</span>}
            hint="คน"
          />
        )}
        {hasTimetable && (
          <Kpi
            accent="blue"
            icon={<ScheduleOutlined />}
            label="คาบสอนวันนี้"
            value={<span className="num">{todayLessons}</span>}
            hint="คาบ"
          />
        )}
        {hasTimetable && (
          <Kpi
            accent="violet"
            icon={<CalendarOutlined />}
            label="คาบสอน/สัปดาห์"
            value={<span className="num">{weekLessons}</span>}
            hint="คาบ"
          />
        )}
        {isAdvisor && (
          <Kpi
            accent="amber"
            icon={<ClockCircleOutlined />}
            label="ยังไม่เช็คชื่อวันนี้"
            value={<span className="num">{dashboard?.attendance.unchecked ?? 0}</span>}
            hint={`จากนักเรียน ${dashboard?.advisee_count ?? 0} คน`}
          />
        )}
      </div>

      {/* เมนูการทำงาน + สถานะบัญชี ข้างกัน (ระหว่าง KPI กับสรุปครู) */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <SectionCard
          className="lg:col-span-2"
          icon={<AppstoreOutlined />}
          title="เมนูการทำงาน"
          description="เข้าถึงงานตามสิทธิ์ของคุณ"
        >
          <div className="flex flex-wrap gap-2.5">
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
                    <span className="text-sm font-medium text-slate-500">{item.label}</span>
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
                  <span className="text-sm font-semibold text-slate-800">{item.label}</span>
                  <ArrowRightOutlined className="text-slate-300 transition-all group-hover:translate-x-0.5 group-hover:text-brand" />
                </button>
              );
            })}
          </div>
        </SectionCard>

        {/* rail ขวา: สถานะบัญชี + บุคลากรล่าสุด */}
        <div className="flex flex-col gap-4">
          {/* การ์ดสถานะบัญชี — รวมบทบาท สิทธิ์ และกลุ่มงานที่สังกัด */}
          <SectionCard icon={<SafetyOutlined />} title="สถานะบัญชี" accent="emerald">
            <dl className="flex flex-col gap-3 text-sm">
              <div className="flex items-center justify-between">
                <dt className="text-slate-500">บทบาท</dt>
                <dd className="font-semibold text-slate-800">{ROLE_LABEL[role]}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt className="text-slate-500">สิทธิ์</dt>
                <dd>
                  <Tag color={user.is_school_admin ? "gold" : "default"} bordered={false}>
                    {user.is_school_admin ? "ผู้ดูแลระบบ" : "สมาชิก"}
                  </Tag>
                </dd>
              </div>
              <div className="flex flex-col gap-1.5">
                <dt className="text-slate-500">กลุ่มงานที่สังกัด</dt>
                <dd>
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
                    <span className="text-slate-400">ยังไม่ได้สังกัดกลุ่มงาน — โปรดให้ผู้ดูแลมอบหมาย</span>
                  )}
                </dd>
              </div>
            </dl>
          </SectionCard>

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
        </div>
      </div>

      {/* สรุปสำหรับครู: ที่ปรึกษา + ตารางสอนวันนี้ (แสดงเมื่อมีข้อมูล) */}
      <HomeTeacherSummary />
    </div>
  );
}
