"use client";

import { CheckCircleFilled, ExclamationCircleOutlined } from "@ant-design/icons";
import { App, Button, DatePicker, Drawer, Select, Spin } from "antd";
import dayjs from "dayjs";
import { useEffect, useMemo, useState, type ReactNode } from "react";

import { AttendanceRadioTable, type MarkState } from "@/features/attendance/AttendanceRadioTable";
import { isAttendanceStatus, type AttendanceMarkBody } from "@/shared/schemas/attendance";
import { DAY_LABELS, type CheckinSlot } from "@/shared/schemas/timetable";

import { useCheckinOverview, useSaveSubjectAttendance, useSubjectRoster } from "./hooks";

// ===== Drawer เช็คชื่อของคาบที่เลือก (ตารางเรดิโอคอลัมน์) =====
function CheckinDrawer(props: { slot: CheckinSlot; onClose: () => void }): ReactNode {
  const { slot, onClose } = props;
  const { message } = App.useApp();
  const { data: roster, isLoading } = useSubjectRoster(slot.slot_id, slot.date);
  const saveMutation = useSaveSubjectAttendance(slot.slot_id);
  const [marks, setMarks] = useState<Record<string, MarkState>>({});

  useEffect(() => {
    if (!roster) return;
    const next: Record<string, MarkState> = {};
    for (const r of roster) next[r.student_id] = { status: r.status, note: r.note };
    setMarks(next);
  }, [roster]);

  const setMark = (id: string, patch: Partial<MarkState>): void => {
    setMarks((prev) => ({
      ...prev,
      [id]: { status: prev[id]?.status ?? "", note: prev[id]?.note ?? "", ...patch },
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
    for (const [id, m] of Object.entries(marks)) {
      if (isAttendanceStatus(m.status)) records.push({ student_id: id, status: m.status, note: m.note });
    }
    if (records.length === 0) {
      message.warning("ยังไม่ได้เลือกสถานะให้นักเรียนคนใด");
      return;
    }
    saveMutation.mutate(
      { date: slot.date, records },
      {
        onSuccess: () => {
          message.success(`บันทึกการเช็คชื่อแล้ว (${records.length} คน)`);
          onClose();
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  return (
    <Drawer
      open
      width={720}
      onClose={onClose}
      title={
        <div>
          <div className="text-sm font-semibold text-slate-800">
            {slot.subject_code} {slot.subject_name}
          </div>
          <div className="text-xs font-normal text-slate-400">
            {slot.class_label} · {DAY_LABELS[slot.day_of_week]} คาบ {slot.period_no} ·{" "}
            <span className="num">{slot.date}</span>
          </div>
        </div>
      }
      extra={
        <div className="flex gap-2">
          <Button onClick={markAllPresent} disabled={!roster || roster.length === 0}>
            ทั้งหมดมาเรียน
          </Button>
          <Button type="primary" loading={saveMutation.isPending} onClick={handleSave} disabled={!roster || roster.length === 0}>
            บันทึก
          </Button>
        </div>
      }
    >
      <AttendanceRadioTable roster={roster ?? []} marks={marks} onMark={setMark} loading={isLoading} />
    </Drawer>
  );
}

const TH_MON = ["", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."];
function thaiDate(iso: string): string {
  const d = dayjs(iso);
  return `${d.date()} ${TH_MON[d.month() + 1]}`;
}

export function SubjectAttendanceTab(): ReactNode {
  const [date, setDate] = useState(dayjs().format("YYYY-MM-DD"));
  const { data, isLoading } = useCheckinOverview(date);
  const [selected, setSelected] = useState<CheckinSlot | null>(null);

  // ถ้าวันที่ปัจจุบันอยู่นอกช่วงเทอม แต่มีรายการสัปดาห์ → snap ไปสัปดาห์แรก
  useEffect(() => {
    if (data && data.weeks.length > 0 && data.current_week_index === 0) {
      setDate(data.weeks[0]!.start);
    }
  }, [data]);

  // มิติกริดจากคาบที่ครูสอน
  const { days, periods, slotMap } = useMemo(() => {
    const slots = data?.slots ?? [];
    let maxDay = 5;
    let maxPeriod = 1;
    const m = new Map<string, CheckinSlot>();
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
  }, [data]);

  const hasSlots = (data?.slots.length ?? 0) > 0;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">สัปดาห์</label>
          {data && data.weeks.length > 0 ? (
            <Select
              style={{ width: 280 }}
              value={data.current_week_index > 0 ? data.week_start : undefined}
              placeholder="เลือกสัปดาห์"
              onChange={(v) => setDate(v)}
              options={data.weeks.map((w) => ({
                value: w.start,
                label: `สัปดาห์ที่ ${w.index} (${thaiDate(w.start)}–${thaiDate(w.end)})`,
              }))}
            />
          ) : (
            // ยังไม่ได้ตั้งวันเปิด-ปิดเทอม → ใช้เลือกวันที่แทน
            <DatePicker
              format="YYYY-MM-DD"
              allowClear={false}
              value={dayjs(date)}
              onChange={(d) => setDate(d ? d.format("YYYY-MM-DD") : dayjs().format("YYYY-MM-DD"))}
            />
          )}
        </div>
        {data && (
          <div className="flex flex-wrap gap-2 text-xs">
            <span className="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-slate-600">
              สัปดาห์นี้ยังไม่เช็ค{" "}
              <span className="num font-semibold text-amber-600">{data.unchecked_this_week}</span>/
              <span className="num">{data.total_this_week}</span> คาบ
            </span>
            {data.has_week_stats && (
              <span className="rounded-lg border border-amber-100 bg-amber-50 px-2.5 py-1.5 text-amber-700">
                ทั้งเทอม: เช็คยังไม่ครบ{" "}
                <span className="num font-semibold">{data.incomplete_weeks}</span>/
                <span className="num">{data.total_weeks}</span> สัปดาห์
              </span>
            )}
          </div>
        )}
      </div>

      {isLoading ? (
        <Spin />
      ) : !hasSlots ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-10 text-center text-sm text-slate-400">
          ไม่มีคาบสอนของคุณในเทอมนี้ — มอบหมายการสอน/จัดตารางที่เมนูตารางสอนก่อน
        </div>
      ) : (
        <>
          <p className="text-xs text-slate-400">คลิกคาบเพื่อเช็คชื่อ · คาบสีเหลือง = ยังไม่เช็คในสัปดาห์นี้</p>
          <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
            <table className="w-full border-collapse text-center text-xs">
              <thead>
                <tr className="bg-slate-50 text-slate-500">
                  <th className="border border-slate-100 px-2 py-2 font-medium">คาบ / วัน</th>
                  {days.map((d) => (
                    <th key={d} className="border border-slate-100 px-2 py-2 font-medium">
                      {DAY_LABELS[d]}
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
                      if (!slot) {
                        return <td key={day} className="border border-slate-100 p-1" />;
                      }
                      return (
                        <td key={day} className="border border-slate-100 p-1 align-top">
                          <button
                            type="button"
                            onClick={() => setSelected(slot)}
                            className={`w-full rounded-lg px-2 py-1.5 text-left transition-colors ${
                              slot.checked
                                ? "bg-emerald-50 hover:bg-emerald-100"
                                : "bg-amber-50 ring-1 ring-amber-200 hover:bg-amber-100"
                            }`}
                          >
                            <div className="flex items-center justify-between gap-1">
                              <span className="num text-[11px] font-medium text-slate-700">{slot.subject_code}</span>
                              {slot.checked ? (
                                <CheckCircleFilled className="text-[11px] text-emerald-500" />
                              ) : (
                                <ExclamationCircleOutlined className="text-[11px] text-amber-500" />
                              )}
                            </div>
                            <div className="truncate text-[11px] text-slate-500">{slot.class_label}</div>
                          </button>
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {selected && <CheckinDrawer slot={selected} onClose={() => setSelected(null)} />}
    </div>
  );
}
