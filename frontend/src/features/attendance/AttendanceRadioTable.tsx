"use client";

import { Input, Radio, Spin, Tag } from "antd";
import type { ReactNode } from "react";

import type { AttendanceRosterEntry, AttendanceStatus } from "@/shared/schemas/attendance";

export type MarkState = { status: string; note: string };

// คอลัมน์สถานะแบบเรดิโอ (วงกลม) — ลำดับตามที่ผู้ใช้ต้องการ: มา/สาย/ขาด/ลากิจ/ลาป่วย
const STATUS_COLUMNS: { value: AttendanceStatus; label: string; color: string }[] = [
  { value: "present", label: "มา", color: "text-emerald-600" },
  { value: "late", label: "สาย", color: "text-amber-600" },
  { value: "absent", label: "ขาด", color: "text-red-600" },
  { value: "personal_leave", label: "ลากิจ", color: "text-violet-600" },
  { value: "sick_leave", label: "ลาป่วย", color: "text-sky-600" },
];

// ตารางเช็คชื่อแบบเรดิโอคอลัมน์ — แต่ละสถานะเป็นคอลัมน์ ติ๊กวงกลมในคอลัมน์เพื่อเลือก
export function AttendanceRadioTable(props: {
  roster: AttendanceRosterEntry[];
  marks: Record<string, MarkState>;
  onMark: (studentId: string, patch: Partial<MarkState>) => void;
  loading?: boolean;
  showNote?: boolean;
}): ReactNode {
  const { roster, marks, onMark, loading, showNote = true } = props;

  if (loading) {
    return (
      <div className="flex justify-center py-10">
        <Spin />
      </div>
    );
  }
  if (roster.length === 0) {
    return <div className="py-8 text-center text-sm text-slate-400">ไม่มีนักเรียนในห้องนี้</div>;
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-100 bg-slate-50 text-xs text-slate-500">
            <th className="px-3 py-2 text-left">เลขที่</th>
            <th className="px-3 py-2 text-left">นักเรียน</th>
            {STATUS_COLUMNS.map((c) => (
              <th key={c.value} className={`px-2 py-2 text-center font-medium ${c.color}`}>
                {c.label}
              </th>
            ))}
            {showNote && <th className="px-3 py-2 text-left">หมายเหตุ</th>}
          </tr>
        </thead>
        <tbody>
          {roster.map((r) => {
            const current = marks[r.student_id]?.status ?? "";
            return (
              <tr key={r.student_id} className="border-b border-slate-50">
                <td className="num px-3 py-2 text-slate-500">{r.student_no ?? "—"}</td>
                <td className="px-3 py-2">
                  <div className="font-medium text-slate-800">
                    {r.prefix}
                    {r.first_name} {r.last_name}
                    {r.daily_status === "late" && (
                      <Tag color="warning" bordered={false} className="ml-1.5 align-middle">
                        มาสาย (เข้าเรียน)
                      </Tag>
                    )}
                  </div>
                  <div className="num text-xs text-slate-400">{r.student_code}</div>
                </td>
                {STATUS_COLUMNS.map((c) => (
                  <td key={c.value} className="px-2 py-2 text-center">
                    <Radio
                      checked={current === c.value}
                      onChange={() => onMark(r.student_id, { status: c.value })}
                      aria-label={`${r.first_name} ${c.label}`}
                    />
                  </td>
                ))}
                {showNote && (
                  <td className="px-3 py-2">
                    <Input
                      size="small"
                      value={marks[r.student_id]?.note ?? ""}
                      onChange={(e) => onMark(r.student_id, { note: e.target.value })}
                      placeholder="ไม่บังคับ"
                    />
                  </td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
