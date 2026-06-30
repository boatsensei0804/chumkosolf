"use client";

import { CloseOutlined, HolderOutlined } from "@ant-design/icons";
import { App, Empty, Select, Spin } from "antd";
import { useMemo, useState, type DragEvent, type ReactNode } from "react";

import { useClassList } from "@/features/classes/hooks";
import { DAY_LABELS, type PeriodDef, type TeachingAssignment, type TimetableSlot } from "@/shared/schemas/timetable";

import {
  useAssignments,
  useClearSlot,
  useSetSlot,
  useSlots,
  useTimetableConfig,
} from "./hooks";

// ข้อมูลที่ลากอยู่: วิชาที่จะวาง + ช่องต้นทาง (ถ้าเป็นการย้าย)
type DragPayload = { assignmentId: string; fromSlotId: string };

const MIME = "application/x-chumkosoft-assignment";

export function GridTab({ readOnly = false }: { readOnly?: boolean }): ReactNode {
  const { message } = App.useApp();
  const { data: classes } = useClassList();
  const { data: config, isLoading: loadingConfig } = useTimetableConfig();
  const { data: assignments } = useAssignments({ enabled: !readOnly });
  const [classId, setClassId] = useState("");
  const { data: slots, isLoading: loadingSlots } = useSlots(classId);
  const setSlotMutation = useSetSlot(classId);
  const clearSlotMutation = useClearSlot(classId);

  const [dragOverKey, setDragOverKey] = useState("");

  const slotMap = useMemo(() => {
    const m = new Map<string, TimetableSlot>();
    for (const s of slots ?? []) m.set(`${s.day_of_week}-${s.period_no}`, s);
    return m;
  }, [slots]);

  const classAssignments = (assignments ?? []).filter((a) => a.class_id === classId);

  // จำนวนคาบที่จัดไปแล้วของแต่ละวิชา (ใช้แสดงใน palette)
  const placedCount = useMemo(() => {
    const m = new Map<string, number>();
    for (const s of slots ?? []) m.set(s.teaching_assignment_id, (m.get(s.teaching_assignment_id) ?? 0) + 1);
    return m;
  }, [slots]);

  const days = config?.days_per_week ?? 5;
  const periodsCount = config?.periods_per_day ?? 8;
  const periodDef = (no: number): PeriodDef | undefined => config?.periods.find((p) => p.period_no === no);
  const periodLabel = (no: number): string => periodDef(no)?.label || `คาบ ${no}`;

  // --- วางวิชาลงช่อง (วางใหม่ หรือ ย้ายจากช่องเดิม) ---
  const placeAt = (day: number, period: number, payload: DragPayload): void => {
    setSlotMutation.mutate(
      { day_of_week: day, period_no: period, teaching_assignment_id: payload.assignmentId },
      {
        onSuccess: () => {
          // ถ้าเป็นการย้าย (มีต้นทาง และไม่ใช่ช่องเดิม) → ล้างช่องต้นทาง
          if (payload.fromSlotId !== "" && slotMap.get(`${day}-${period}`)?.id !== payload.fromSlotId) {
            clearSlotMutation.mutate(payload.fromSlotId);
          }
          message.success("จัดคาบแล้ว");
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  const onDrop = (e: DragEvent<HTMLDivElement>, day: number, period: number): void => {
    e.preventDefault();
    setDragOverKey("");
    const raw = e.dataTransfer.getData(MIME);
    if (!raw) return;
    try {
      const payload = JSON.parse(raw) as DragPayload;
      if (payload.assignmentId) placeAt(day, period, payload);
    } catch {
      /* ignore payload ที่อ่านไม่ได้ */
    }
  };

  const allowDrop = (e: DragEvent<HTMLDivElement>, key: string): void => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    if (dragOverKey !== key) setDragOverKey(key);
  };

  const startDrag = (e: DragEvent<HTMLElement>, payload: DragPayload): void => {
    e.dataTransfer.setData(MIME, JSON.stringify(payload));
    e.dataTransfer.effectAllowed = "move";
  };

  const clearCell = (slot: TimetableSlot): void => {
    clearSlotMutation.mutate(slot.id, {
      onSuccess: () => message.success("ล้างช่องแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ห้องเรียน</label>
          <Select
            showSearch
            value={classId || undefined}
            onChange={setClassId}
            placeholder={readOnly ? "เลือกห้องเพื่อดูตาราง" : "เลือกห้องเพื่อจัดตาราง"}
            style={{ width: 240 }}
            optionFilterProp="label"
            options={(classes ?? []).map((c) => ({ value: c.id, label: `${c.grade_level} ${c.room_name}` }))}
          />
        </div>
        {classId !== "" && !readOnly && (
          <p className="text-xs text-slate-400">
            ลากวิชาจากด้านซ้ายมาวางในช่อง · ลากช่องเพื่อย้าย · กดกากบาทเพื่อลบ
          </p>
        )}
      </div>

      {classId === "" ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-10 text-center text-sm text-slate-400">
          เลือกห้องเพื่อจัดตารางสอน
        </div>
      ) : loadingConfig || loadingSlots ? (
        <Spin />
      ) : (
        <div className={readOnly ? "" : "grid grid-cols-1 gap-4 lg:grid-cols-[220px_1fr]"}>
          {/* palette วิชาของห้องนี้ (ลากได้) — เฉพาะโหมดแก้ไข */}
          {!readOnly && (
            <AssignmentPalette assignments={classAssignments} placedCount={placedCount} onDragStart={startDrag} />
          )}

          {/* ตารางห้องเรียน */}
          <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
            <table className="w-full border-collapse text-center text-xs">
              <thead>
                <tr className="bg-slate-50 text-slate-500">
                  <th className="border border-slate-100 px-2 py-2 font-medium">คาบ / วัน</th>
                  {Array.from({ length: days }, (_, i) => i + 1).map((d) => (
                    <th key={d} className="border border-slate-100 px-2 py-2 font-medium">
                      {DAY_LABELS[d]}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {Array.from({ length: periodsCount }, (_, i) => i + 1).map((period) => {
                  const def = periodDef(period);
                  const time = def && def.start_time !== "" ? `${def.start_time}-${def.end_time}` : "";
                  // แถวคาบพัก — วางไม่ได้
                  if (def?.is_break) {
                    return (
                      <tr key={period}>
                        <td className="border border-slate-100 bg-amber-50/60 px-2 py-2 text-left text-slate-500">
                          <div>{periodLabel(period)}</div>
                          {time !== "" && <div className="num text-[10px] text-slate-400">{time}</div>}
                        </td>
                        <td colSpan={days} className="border border-slate-100 bg-amber-50/60 px-2 py-2 text-amber-600">
                          พัก
                        </td>
                      </tr>
                    );
                  }
                  return (
                    <tr key={period}>
                      <td className="border border-slate-100 bg-slate-50 px-2 py-2 text-left text-slate-500">
                        <div>{periodLabel(period)}</div>
                        {time !== "" && <div className="num text-[10px] text-slate-400">{time}</div>}
                      </td>
                      {Array.from({ length: days }, (_, i) => i + 1).map((day) => {
                        const key = `${day}-${period}`;
                        const slot = slotMap.get(key);
                        return (
                          <td
                            key={day}
                            className={`border border-slate-100 p-1 align-top transition-colors ${dragOverKey === key ? "bg-sky-50" : ""}`}
                            {...(readOnly
                              ? {}
                              : {
                                  onDragOver: (e: DragEvent<HTMLTableCellElement>) => allowDrop(e, key),
                                  onDragLeave: () => dragOverKey === key && setDragOverKey(""),
                                  onDrop: (e: DragEvent<HTMLTableCellElement>) => onDrop(e, day, period),
                                })}
                          >
                            {slot ? (
                              <div
                                {...(readOnly
                                  ? {}
                                  : {
                                      draggable: true,
                                      onDragStart: (e: DragEvent<HTMLDivElement>) =>
                                        startDrag(e, { assignmentId: slot.teaching_assignment_id, fromSlotId: slot.id }),
                                    })}
                                className={`group relative rounded-lg bg-sky-50 px-2 py-1.5 text-left ${readOnly ? "" : "cursor-grab active:cursor-grabbing"}`}
                              >
                                <div className="num text-[11px] font-medium text-sky-700">{slot.subject_code}</div>
                                <div className="truncate text-[11px] text-slate-500">{slot.teacher_name}</div>
                                {!readOnly && (
                                  <button
                                    type="button"
                                    aria-label="ล้างช่อง"
                                    onClick={() => clearCell(slot)}
                                    className="absolute right-0.5 top-0.5 hidden rounded text-slate-400 hover:text-red-500 group-hover:block"
                                  >
                                    <CloseOutlined />
                                  </button>
                                )}
                              </div>
                            ) : (
                              <div className="flex h-10 w-full items-center justify-center rounded-lg text-[10px] text-slate-200">
                                ว่าง
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

// ===== palette รายการวิชาของห้อง (ลากได้) + จำนวนคาบที่จัดแล้ว =====
function AssignmentPalette(props: {
  assignments: TeachingAssignment[];
  placedCount: Map<string, number>;
  onDragStart: (e: DragEvent<HTMLElement>, payload: DragPayload) => void;
}): ReactNode {
  const { assignments, placedCount, onDragStart } = props;
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <div className="mb-2 text-xs font-medium text-slate-500">วิชาของห้องนี้</div>
      {assignments.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={<span className="text-xs text-slate-400">ยังไม่มีการมอบหมายสอน — เพิ่มที่แท็บ “มอบหมายสอน”</span>}
        />
      ) : (
        <ul className="flex flex-col gap-2">
          {assignments.map((a) => (
            <li
              key={a.id}
              draggable
              onDragStart={(e) => onDragStart(e, { assignmentId: a.id, fromSlotId: "" })}
              className="cursor-grab rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-2 active:cursor-grabbing"
            >
              <div className="flex items-center gap-1.5">
                <HolderOutlined className="text-slate-300" />
                <span className="num text-[11px] font-medium text-sky-700">{a.subject_code}</span>
                <span className="ml-auto rounded bg-white px-1.5 text-[10px] text-slate-400">
                  จัดแล้ว <span className="num">{placedCount.get(a.id) ?? 0}</span>
                </span>
              </div>
              <div className="truncate text-[11px] text-slate-600">{a.subject_name}</div>
              <div className="truncate text-[10px] text-slate-400">{a.teacher_name}</div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
