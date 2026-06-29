"use client";

import { App, Button, Checkbox, Input, InputNumber, Spin, TimePicker } from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { useEffect, useState, type ReactNode } from "react";

import type { ConfigBody, PeriodDef } from "@/shared/schemas/timetable";

import { useSaveConfig, useTimetableConfig } from "./hooks";

function hmToDayjs(s: string): Dayjs | null {
  if (!/^\d{2}:\d{2}$/.test(s)) return null;
  const parts = s.split(":");
  return dayjs().hour(Number(parts[0])).minute(Number(parts[1])).second(0);
}

// ปรับจำนวนแถวคาบให้เท่ากับ periodsPerDay (คงค่าที่กรอกไว้เดิม)
function resizePeriods(existing: PeriodDef[], count: number): PeriodDef[] {
  const out: PeriodDef[] = [];
  for (let n = 1; n <= count; n++) {
    const prev = existing.find((p) => p.period_no === n);
    out.push(prev ?? { period_no: n, label: `คาบ ${n}`, start_time: "", end_time: "", is_break: false });
  }
  return out;
}

export function ConfigTab(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useTimetableConfig();
  const saveMutation = useSaveConfig();

  const [daysPerWeek, setDaysPerWeek] = useState(5);
  const [periodsPerDay, setPeriodsPerDay] = useState(8);
  const [periods, setPeriods] = useState<PeriodDef[]>([]);

  useEffect(() => {
    if (!data) return;
    setDaysPerWeek(data.days_per_week);
    setPeriodsPerDay(data.periods_per_day);
    setPeriods(resizePeriods(data.periods, data.periods_per_day));
  }, [data]);

  const changePeriodsPerDay = (v: number | null): void => {
    const n = v ?? 1;
    setPeriodsPerDay(n);
    setPeriods((prev) => resizePeriods(prev, n));
  };

  const updateRow = (no: number, patch: Partial<PeriodDef>): void => {
    setPeriods((prev) => prev.map((p) => (p.period_no === no ? { ...p, ...patch } : p)));
  };

  const handleSave = (): void => {
    const body: ConfigBody = { days_per_week: daysPerWeek, periods_per_day: periodsPerDay, periods };
    saveMutation.mutate(body, {
      onSuccess: () => message.success("บันทึกการตั้งค่าตารางสอนแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  if (isLoading) return <Spin />;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end gap-4">
        <div>
          <label className="mb-1 block text-xs text-slate-500">จำนวนวันเรียน/สัปดาห์</label>
          <InputNumber min={1} max={7} value={daysPerWeek} onChange={(v) => setDaysPerWeek(v ?? 1)} style={{ width: 90 }} />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">จำนวนคาบ/วัน</label>
          <InputNumber min={1} max={20} value={periodsPerDay} onChange={changePeriodsPerDay} style={{ width: 90 }} />
        </div>
        <Button type="primary" loading={saveMutation.isPending} onClick={handleSave}>
          บันทึกการตั้งค่า
        </Button>
      </div>

      <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-100 text-left text-xs text-slate-400">
              <th className="px-3 py-2">คาบ</th>
              <th className="px-3 py-2">ชื่อคาบ</th>
              <th className="px-3 py-2">เริ่ม</th>
              <th className="px-3 py-2">จบ</th>
              <th className="px-3 py-2">พัก</th>
            </tr>
          </thead>
          <tbody>
            {periods.map((p) => (
              <tr key={p.period_no} className="border-b border-slate-50">
                <td className="px-3 py-2 num text-slate-500">{p.period_no}</td>
                <td className="px-3 py-2">
                  <Input
                    size="small"
                    value={p.label}
                    onChange={(e) => updateRow(p.period_no, { label: e.target.value })}
                    placeholder={`คาบ ${p.period_no}`}
                    style={{ width: 160 }}
                  />
                </td>
                <td className="px-3 py-2">
                  <TimePicker
                    size="small"
                    format="HH:mm"
                    minuteStep={5}
                    value={hmToDayjs(p.start_time)}
                    onChange={(d) => updateRow(p.period_no, { start_time: d ? d.format("HH:mm") : "" })}
                  />
                </td>
                <td className="px-3 py-2">
                  <TimePicker
                    size="small"
                    format="HH:mm"
                    minuteStep={5}
                    value={hmToDayjs(p.end_time)}
                    onChange={(d) => updateRow(p.period_no, { end_time: d ? d.format("HH:mm") : "" })}
                  />
                </td>
                <td className="px-3 py-2">
                  <Checkbox
                    checked={p.is_break}
                    onChange={(e) => updateRow(p.period_no, { is_break: e.target.checked })}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
