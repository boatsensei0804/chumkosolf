"use client";

import { Empty, Segmented, Spin, Tag } from "antd";
import { useState, type ReactNode } from "react";

import { DAY_LABELS } from "@/shared/schemas/timetable";

import { useFreeTeachers } from "./hooks";

function todayWeekday(): number {
  const wd = new Date().getDay(); // อาทิตย์=0
  return wd === 0 ? 7 : wd;
}

// FreeTeachersTab — แสดงครูที่ว่าง (ไม่ติดสอน) ในแต่ละคาบของวันที่เลือก (read-only)
export function FreeTeachersTab(): ReactNode {
  const [day, setDay] = useState<number>(todayWeekday());
  const { data, isLoading, isError, error } = useFreeTeachers(day);

  const dayOptions = [1, 2, 3, 4, 5, 6, 7].map((d) => ({ label: DAY_LABELS[d], value: d }));

  return (
    <div className="flex flex-col gap-4 pt-2">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-slate-500">เลือกวันเพื่อดูว่าครูท่านใดว่าง (ไม่มีคาบสอน) ในแต่ละคาบ</p>
        <Segmented
          options={dayOptions}
          value={day}
          onChange={(v) => setDay(Number(v))}
          aria-label="เลือกวัน"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spin />
        </div>
      ) : isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          โหลดข้อมูลไม่สำเร็จ: {error?.message}
        </div>
      ) : !data || data.periods.length === 0 ? (
        <Empty description="ยังไม่ได้ตั้งค่าคาบเรียนของเทอมนี้" />
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {data.periods.map((p) => (
            <div key={p.period_no} className="overflow-hidden rounded-xl border border-slate-200 bg-white">
              <div className="flex items-center justify-between border-b border-slate-100 bg-slate-50 px-3 py-2">
                <span className="text-sm font-semibold text-slate-700">
                  คาบ {p.period_no}
                  {p.label ? ` · ${p.label}` : ""}
                </span>
                <span className="num text-xs text-slate-400">ว่าง {p.free_teachers.length}</span>
              </div>
              <div className="flex flex-wrap gap-1.5 p-3">
                {p.free_teachers.length === 0 ? (
                  <span className="text-sm text-slate-400">ไม่มีครูว่างในคาบนี้</span>
                ) : (
                  p.free_teachers.map((t) => (
                    <Tag key={t.id} bordered={false} color="green">
                      {t.name}
                    </Tag>
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
