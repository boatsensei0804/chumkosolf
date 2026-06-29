"use client";

import { App, Button, DatePicker, Select } from "antd";
import dayjs from "dayjs";
import { useEffect, useState, type ReactNode } from "react";

import { useClassList } from "@/features/classes/hooks";
import { isAttendanceStatus, type AttendanceMarkBody } from "@/shared/schemas/attendance";

import { AttendanceRadioTable, type MarkState } from "./AttendanceRadioTable";
import { useRoster, useSaveAttendance } from "./hooks";

export function AttendanceTab(): ReactNode {
  const { message } = App.useApp();
  const { data: classes, isLoading: loadingClasses } = useClassList();
  const [classId, setClassId] = useState("");
  const [date, setDate] = useState(dayjs().format("YYYY-MM-DD"));

  const { data: roster, isLoading: loadingRoster, isError, error } = useRoster(classId, date);
  const saveMutation = useSaveAttendance(classId);

  // marks: สถานะที่ผู้ใช้กำลังแก้ ของแต่ละนักเรียน (sync จาก roster เมื่อโหลดใหม่)
  const [marks, setMarks] = useState<Record<string, MarkState>>({});
  useEffect(() => {
    if (!roster) return;
    const next: Record<string, MarkState> = {};
    for (const r of roster) next[r.student_id] = { status: r.status, note: r.note };
    setMarks(next);
  }, [roster]);

  const setMark = (studentId: string, patch: Partial<MarkState>): void => {
    setMarks((prev) => ({
      ...prev,
      [studentId]: { status: prev[studentId]?.status ?? "", note: prev[studentId]?.note ?? "", ...patch },
    }));
  };

  const markAllPresent = (): void => {
    if (!roster) return;
    const next: Record<string, MarkState> = {};
    for (const r of roster) next[r.student_id] = { status: "present", note: marks[r.student_id]?.note ?? "" };
    setMarks(next);
  };

  const handleSave = (): void => {
    const records: AttendanceMarkBody[] = [];
    for (const [studentId, m] of Object.entries(marks)) {
      if (isAttendanceStatus(m.status)) {
        records.push({ student_id: studentId, status: m.status, note: m.note });
      }
    }
    if (records.length === 0) {
      message.warning("ยังไม่ได้เลือกสถานะให้นักเรียนคนใด");
      return;
    }
    saveMutation.mutate(
      { date, records },
      {
        onSuccess: () => message.success(`บันทึกการเช็คชื่อแล้ว (${records.length} คน)`),
        onError: (err) => message.error(err.message),
      },
    );
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ห้องเรียน</label>
          <Select
            showSearch
            value={classId || undefined}
            onChange={setClassId}
            loading={loadingClasses}
            placeholder="เลือกห้อง"
            style={{ width: 220 }}
            optionFilterProp="label"
            options={(classes ?? []).map((c) => ({
              value: c.id,
              label: `${c.grade_level} ${c.room_name}`,
            }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วันที่</label>
          <DatePicker
            format="YYYY-MM-DD"
            allowClear={false}
            value={dayjs(date)}
            onChange={(d) => setDate(d ? d.format("YYYY-MM-DD") : dayjs().format("YYYY-MM-DD"))}
          />
        </div>
        <Button onClick={markAllPresent} disabled={!roster || roster.length === 0}>
          ทั้งหมดมาเรียน
        </Button>
        <Button
          type="primary"
          loading={saveMutation.isPending}
          disabled={!classId || !roster || roster.length === 0}
          onClick={handleSave}
        >
          บันทึกการเช็คชื่อ
        </Button>
      </div>

      {classId === "" ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-10 text-center text-sm text-slate-400">
          เลือกห้องเรียนเพื่อเริ่มเช็คชื่อ
        </div>
      ) : isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          โหลดรายชื่อไม่สำเร็จ: {error?.message}
        </div>
      ) : (
        <AttendanceRadioTable roster={roster ?? []} marks={marks} onMark={setMark} loading={loadingRoster} />
      )}
    </div>
  );
}
