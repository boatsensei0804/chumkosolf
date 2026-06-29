"use client";

import { CalendarOutlined, TeamOutlined } from "@ant-design/icons";
import { useMemo, type ReactNode } from "react";

import { SectionCard } from "@/shared/ui/SectionCard";
import type { DashboardSlot } from "@/shared/schemas/dashboard";
import { DAY_LABELS } from "@/shared/schemas/timetable";

import { useDashboard } from "./hooks";

function Stat({ label, value, color }: { label: string; value: number; color: string }): ReactNode {
  return (
    <div className="rounded-lg border border-slate-100 bg-white px-3 py-2 text-center">
      <div className={`num text-xl font-bold leading-none ${color}`}>{value}</div>
      <div className="mt-1 text-[11px] text-slate-400">{label}</div>
    </div>
  );
}

// กริดตารางสอนของครู ไฮไลต์คอลัมน์ของวันนี้
function MiniTimetable({ slots, today }: { slots: DashboardSlot[]; today: number }): ReactNode {
  const { days, periods, slotMap } = useMemo(() => {
    let maxDay = 5;
    let maxPeriod = 1;
    const m = new Map<string, DashboardSlot>();
    for (const s of slots) {
      maxDay = Math.max(maxDay, s.day_of_week);
      maxPeriod = Math.max(maxPeriod, s.period_no);
      m.set(`${s.day_of_week}-${s.period_no}`, s);
    }
    return {
      days: Array.from({ length: maxDay }, (_, i) => i + 1),
      periods: Array.from({ length: maxPeriod }, (_, i) => i + 1),
      slotMap: m,
    };
  }, [slots]);

  return (
    <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
      <table className="w-full border-collapse text-center text-xs">
        <thead>
          <tr className="bg-slate-50 text-slate-500">
            <th className="border border-slate-100 px-2 py-2 font-medium">คาบ / วัน</th>
            {days.map((d) => (
              <th
                key={d}
                className={`border border-slate-100 px-2 py-2 font-medium ${d === today ? "bg-sky-100 text-sky-700" : ""}`}
              >
                {DAY_LABELS[d]}
                {d === today && <span className="block text-[9px] font-normal">วันนี้</span>}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {periods.map((period) => (
            <tr key={period}>
              <td className="border border-slate-100 bg-slate-50 px-2 py-2 text-slate-500">คาบ {period}</td>
              {days.map((day) => {
                const slot = slotMap.get(`${day}-${period}`);
                const isToday = day === today;
                return (
                  <td key={day} className={`border border-slate-100 p-1 ${isToday ? "bg-sky-50" : ""}`}>
                    {slot ? (
                      <div className={`rounded-lg px-2 py-1.5 text-left ${isToday ? "bg-sky-100 ring-1 ring-sky-200" : "bg-slate-50"}`}>
                        <div className="num text-[11px] font-medium text-sky-700">{slot.subject_code}</div>
                        <div className="truncate text-[10px] text-slate-500">{slot.class_label}</div>
                      </div>
                    ) : null}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// HomeTeacherSummary — การ์ดที่ปรึกษา (จำนวน + เช็คชื่อวันนี้) + ตารางสอนไฮไลต์วันนี้
export function HomeTeacherSummary(): ReactNode {
  const { data } = useDashboard();
  if (!data) return null;
  const showAdvisor = data.is_advisor;
  const showTimetable = data.slots.length > 0;
  if (!showAdvisor && !showTimetable) return null;

  const a = data.attendance;

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
      {showAdvisor && (
        <SectionCard
          icon={<TeamOutlined />}
          title="นักเรียนที่ปรึกษา"
          description="สรุปการเช็คชื่อเข้าเรียนของวันนี้"
          accent="emerald"
        >
          <div className="mb-3 flex items-baseline gap-2">
            <span className="num text-3xl font-bold text-brand-navy">{data.advisee_count}</span>
            <span className="text-sm text-slate-400">คน</span>
          </div>
          <div className="grid grid-cols-3 gap-2">
            <Stat label="มา" value={a.present} color="text-emerald-600" />
            <Stat label="สาย" value={a.late} color="text-amber-600" />
            <Stat label="ขาด" value={a.absent} color="text-red-600" />
            <Stat label="ลาป่วย" value={a.sick_leave} color="text-sky-600" />
            <Stat label="ลากิจ" value={a.personal_leave} color="text-violet-600" />
            <Stat label="ยังไม่เช็ค" value={a.unchecked} color="text-slate-400" />
          </div>
        </SectionCard>
      )}

      {showTimetable && (
        <SectionCard
          className={showAdvisor ? "lg:col-span-2" : "lg:col-span-3"}
          icon={<CalendarOutlined />}
          title="ตารางสอนของฉัน"
          description="ไฮไลต์คาบที่สอนในวันนี้"
        >
          <MiniTimetable slots={data.slots} today={data.today_weekday} />
        </SectionCard>
      )}
    </div>
  );
}
