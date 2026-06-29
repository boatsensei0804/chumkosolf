"use client";

import { DeleteOutlined } from "@ant-design/icons";
import {
  App,
  Button,
  DatePicker,
  Input,
  InputNumber,
  Popconfirm,
  Select,
  Spin,
  Tag,
} from "antd";
import dayjs from "dayjs";
import { useState, type ReactNode } from "react";

import { useClassList, useEnrollments } from "@/features/classes/hooks";
import type { BehaviorBody } from "@/shared/schemas/behavior";

import { useAddBehavior, useBehavior, useDeleteBehavior } from "./hooks";

type AdjustType = "deduct" | "add";

// ===== ฟอร์มเพิ่ม/หักคะแนน (presentational) =====
export function BehaviorAddForm(props: {
  onAdd: (body: BehaviorBody) => void;
  submitting: boolean;
}): ReactNode {
  const { onAdd, submitting } = props;
  const [type, setType] = useState<AdjustType>("deduct");
  const [amount, setAmount] = useState<number>(5);
  const [reason, setReason] = useState("");
  const [occurredAt, setOccurredAt] = useState("");
  const [error, setError] = useState("");

  const submit = (): void => {
    if (!amount || amount < 1) {
      setError("กรุณาระบุจำนวนคะแนน (ตั้งแต่ 1)");
      return;
    }
    if (reason.trim() === "") {
      setError("กรุณาระบุเหตุผล");
      return;
    }
    setError("");
    onAdd({
      points: type === "deduct" ? -amount : amount,
      reason: reason.trim(),
      occurred_at: occurredAt,
    });
    setReason("");
    setOccurredAt("");
    setAmount(5);
  };

  return (
    <div className="mt-4 flex flex-col gap-2 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ประเภท</label>
          <Select<AdjustType>
            value={type}
            onChange={setType}
            style={{ width: 110 }}
            options={[
              { value: "deduct", label: "หักคะแนน" },
              { value: "add", label: "เพิ่มคะแนน" },
            ]}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">จำนวน</label>
          <InputNumber
            min={1}
            value={amount}
            onChange={(v) => setAmount(v ?? 0)}
            style={{ width: 90 }}
          />
        </div>
        <div className="min-w-[180px] flex-1">
          <label className="mb-1 block text-xs text-slate-500">เหตุผล</label>
          <Input
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="เช่น มาสาย, ช่วยงานส่วนรวม"
            status={error && reason.trim() === "" ? "error" : ""}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วันที่เกิดเหตุ</label>
          <DatePicker
            format="YYYY-MM-DD"
            value={occurredAt ? dayjs(occurredAt) : null}
            onChange={(d) => setOccurredAt(d ? d.format("YYYY-MM-DD") : "")}
          />
        </div>
        <Button type="primary" loading={submitting} onClick={submit}>
          บันทึก
        </Button>
      </div>
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
}

function scoreColorClass(score: number, start: number): string {
  if (start <= 0) return "text-slate-700";
  if (score >= start * 0.8) return "text-emerald-600";
  if (score >= start * 0.5) return "text-amber-600";
  return "text-red-600";
}

// ===== แผงคะแนนของนักเรียน 1 คน =====
function BehaviorPanel({ studentId }: { studentId: string }): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useBehavior(studentId);
  const addMutation = useAddBehavior(studentId);
  const deleteMutation = useDeleteBehavior(studentId);

  const handleAdd = (body: BehaviorBody): void => {
    addMutation.mutate(body, {
      onSuccess: () => message.success("บันทึกคะแนนแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  if (isLoading) return <Spin />;
  if (!data) return null;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-4">
        <div className="text-center">
          <div className={`num text-3xl font-semibold ${scoreColorClass(data.current_score, data.starting_score)}`}>
            {data.current_score}
          </div>
          <div className="text-xs text-slate-400">คะแนนปัจจุบัน</div>
        </div>
        <div className="text-sm text-slate-500">
          คะแนนตั้งต้น <span className="num">{data.starting_score}</span>
          <br />
          ปรับแล้ว{" "}
          <span className="num">{data.current_score - data.starting_score}</span> คะแนน
        </div>
      </div>

      {data.records.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-8 text-center text-sm text-slate-400">
          ยังไม่มีประวัติการปรับคะแนนในเทอมนี้
        </div>
      ) : (
        <ul className="divide-y divide-slate-100">
          {data.records.map((r) => (
            <li key={r.id} className="flex items-center justify-between gap-3 py-3">
              <div className="flex min-w-0 items-center gap-2">
                <Tag color={r.points < 0 ? "error" : "success"} bordered={false} className="num">
                  {r.points > 0 ? `+${r.points}` : r.points}
                </Tag>
                <span className="truncate text-sm text-slate-700">{r.reason}</span>
                <span className="num text-xs text-slate-400">
                  {r.occurred_at === "" ? "" : r.occurred_at}
                </span>
              </div>
              <Popconfirm
                title="ลบรายการนี้?"
                okText="ลบ"
                cancelText="ยกเลิก"
                okButtonProps={{ danger: true }}
                onConfirm={() =>
                  deleteMutation.mutate(r.id, {
                    onSuccess: () => message.success("ลบรายการแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              >
                <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
              </Popconfirm>
            </li>
          ))}
        </ul>
      )}

      <BehaviorAddForm onAdd={handleAdd} submitting={addMutation.isPending} />
    </div>
  );
}

export function BehaviorTab(): ReactNode {
  const { data: classes, isLoading: loadingClasses } = useClassList();
  const [classId, setClassId] = useState("");
  const { data: students } = useEnrollments(classId);
  const [studentId, setStudentId] = useState("");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ห้องเรียน</label>
          <Select
            showSearch
            value={classId || undefined}
            onChange={(v) => {
              setClassId(v);
              setStudentId("");
            }}
            loading={loadingClasses}
            placeholder="เลือกห้อง"
            style={{ width: 200 }}
            optionFilterProp="label"
            options={(classes ?? []).map((c) => ({
              value: c.id,
              label: `${c.grade_level} ${c.room_name}`,
            }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">นักเรียน</label>
          <Select
            showSearch
            value={studentId || undefined}
            onChange={setStudentId}
            placeholder="เลือกนักเรียน"
            style={{ width: 260 }}
            optionFilterProp="label"
            disabled={classId === ""}
            options={(students ?? []).map((s) => ({
              value: s.student_id,
              label: `${s.prefix}${s.first_name} ${s.last_name}`,
            }))}
          />
        </div>
      </div>

      {studentId === "" ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-10 text-center text-sm text-slate-400">
          เลือกห้องและนักเรียนเพื่อดู/บันทึกคะแนนความประพฤติ
        </div>
      ) : (
        <BehaviorPanel studentId={studentId} />
      )}
    </div>
  );
}
