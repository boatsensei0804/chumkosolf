"use client";

import {
  DeleteOutlined,
  SafetyCertificateOutlined,
  TrophyOutlined,
} from "@ant-design/icons";
import { App, Button, Checkbox, DatePicker, Input, Popconfirm, Select, Spin, Tag } from "antd";
import dayjs from "dayjs";
import { useState, type ReactNode } from "react";

import { SectionCard } from "@/shared/ui/SectionCard";
import {
  positionLabel,
  type CreatePositionBody,
  type StandingBody,
} from "@/shared/schemas/personnelExtra";
import type { AdminPosition } from "@/shared/schemas/enums";

import {
  useCreatePosition,
  useCreateStanding,
  useDeletePosition,
  useDeleteStanding,
  usePositions,
  useStandings,
} from "./extraHooks";

function dateText(s: string): string {
  return s === "" ? "—" : s;
}

// แถบเปล่า/empty ที่สื่อความหมาย
function EmptyHint({ children }: { children: ReactNode }): ReactNode {
  return (
    <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-8 text-center text-sm text-slate-400">
      {children}
    </div>
  );
}

// กรอบสำหรับฟอร์มเพิ่มรายการ
function AddBar({ children }: { children: ReactNode }): ReactNode {
  return (
    <div className="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
      {children}
    </div>
  );
}

// ปุ่มลบรายการ (icon + ยืนยัน)
function DeleteRowButton({ title, onConfirm }: { title: string; onConfirm: () => void }): ReactNode {
  return (
    <Popconfirm
      title={title}
      okText="ลบ"
      cancelText="ยกเลิก"
      okButtonProps={{ danger: true }}
      onConfirm={onConfirm}
    >
      <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
    </Popconfirm>
  );
}

// ===== ฟอร์มเพิ่มตำแหน่งบริหาร (presentational) =====
export function PositionAddForm(props: {
  onAdd: (body: CreatePositionBody) => void;
  submitting: boolean;
}): ReactNode {
  const { onAdd, submitting } = props;
  const [position, setPosition] = useState<AdminPosition>("deputy_director");
  const [appointedAt, setAppointedAt] = useState("");

  return (
    <div className="flex flex-wrap items-end gap-3">
      <div>
        <label className="mb-1 block text-xs text-slate-500">ตำแหน่ง</label>
        <Select<AdminPosition>
          value={position}
          onChange={setPosition}
          style={{ width: 200 }}
          options={[
            { value: "director", label: positionLabel.director },
            { value: "deputy_director", label: positionLabel.deputy_director },
          ]}
        />
      </div>
      <div>
        <label className="mb-1 block text-xs text-slate-500">วันที่แต่งตั้ง</label>
        <DatePicker
          format="YYYY-MM-DD"
          value={appointedAt ? dayjs(appointedAt) : null}
          onChange={(d) => setAppointedAt(d ? d.format("YYYY-MM-DD") : "")}
        />
      </div>
      <Button
        type="primary"
        loading={submitting}
        onClick={() => onAdd({ position, appointed_at: appointedAt })}
      >
        เพิ่มตำแหน่ง
      </Button>
    </div>
  );
}

// ===== ฟอร์มเพิ่มวิทยฐานะ (presentational) =====
export function StandingAddForm(props: {
  onAdd: (body: StandingBody) => void;
  submitting: boolean;
}): ReactNode {
  const { onAdd, submitting } = props;
  const [standing, setStanding] = useState("");
  const [effectiveDate, setEffectiveDate] = useState("");
  const [isCurrent, setIsCurrent] = useState(false);
  const [error, setError] = useState("");

  const submit = (): void => {
    if (standing.trim() === "") {
      setError("กรุณากรอกชื่อวิทยฐานะ");
      return;
    }
    setError("");
    onAdd({ standing: standing.trim(), effective_date: effectiveDate, is_current: isCurrent });
    setStanding("");
    setEffectiveDate("");
    setIsCurrent(false);
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">วิทยฐานะ</label>
          <Input
            value={standing}
            onChange={(e) => setStanding(e.target.value)}
            placeholder="เช่น ครู คศ.1, ชำนาญการ"
            style={{ width: 220 }}
            status={error ? "error" : ""}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วันที่มีผล</label>
          <DatePicker
            format="YYYY-MM-DD"
            value={effectiveDate ? dayjs(effectiveDate) : null}
            onChange={(d) => setEffectiveDate(d ? d.format("YYYY-MM-DD") : "")}
          />
        </div>
        <Checkbox checked={isCurrent} onChange={(e) => setIsCurrent(e.target.checked)}>
          ปัจจุบัน
        </Checkbox>
        <Button type="primary" loading={submitting} onClick={submit}>
          เพิ่มวิทยฐานะ
        </Button>
      </div>
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
}

// ===== ตำแหน่งบริหาร (container) =====
function PositionsSection({ personnelId }: { personnelId: string }): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = usePositions(personnelId);
  const createMutation = useCreatePosition(personnelId);
  const deleteMutation = useDeletePosition(personnelId);

  const handleAdd = (body: CreatePositionBody): void => {
    createMutation.mutate(body, {
      onSuccess: () => message.success("เพิ่มตำแหน่งแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  return (
    <SectionCard
      icon={<SafetyCertificateOutlined />}
      title="ตำแหน่งบริหาร"
      description="ผู้อำนวยการมีได้คนเดียวที่ดำรงตำแหน่ง"
      accent="amber"
    >
      {isLoading ? (
        <Spin />
      ) : (data?.length ?? 0) === 0 ? (
        <EmptyHint>ยังไม่มีตำแหน่งบริหาร</EmptyHint>
      ) : (
        <ul className="divide-y divide-slate-100">
          {(data ?? []).map((p) => (
            <li key={p.id} className="flex items-center justify-between gap-3 py-3">
              <div className="flex flex-wrap items-center gap-2">
                <Tag color="blue" bordered={false}>
                  {p.position === "director" || p.position === "deputy_director"
                    ? positionLabel[p.position]
                    : p.position}
                </Tag>
                {p.is_active && (
                  <Tag color="success" bordered={false}>
                    ดำรงตำแหน่ง
                  </Tag>
                )}
                <span className="text-xs text-slate-400">
                  แต่งตั้ง <span className="num">{dateText(p.appointed_at)}</span>
                </span>
              </div>
              <DeleteRowButton
                title="ลบตำแหน่งนี้?"
                onConfirm={() =>
                  deleteMutation.mutate(p.id, {
                    onSuccess: () => message.success("ลบแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              />
            </li>
          ))}
        </ul>
      )}
      <AddBar>
        <PositionAddForm onAdd={handleAdd} submitting={createMutation.isPending} />
      </AddBar>
    </SectionCard>
  );
}

// ===== วิทยฐานะ (container) =====
function StandingsSection({ personnelId }: { personnelId: string }): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useStandings(personnelId);
  const createMutation = useCreateStanding(personnelId);
  const deleteMutation = useDeleteStanding(personnelId);

  const handleAdd = (body: StandingBody): void => {
    createMutation.mutate(body, {
      onSuccess: () => message.success("เพิ่มวิทยฐานะแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  return (
    <SectionCard
      icon={<TrophyOutlined />}
      title="วิทยฐานะ"
      description="เก็บเป็นประวัติ — ระบุปัจจุบันได้รายการเดียว"
      accent="violet"
    >
      {isLoading ? (
        <Spin />
      ) : (data?.length ?? 0) === 0 ? (
        <EmptyHint>ยังไม่มีประวัติวิทยฐานะ</EmptyHint>
      ) : (
        <ul className="divide-y divide-slate-100">
          {(data ?? []).map((s) => (
            <li key={s.id} className="flex items-center justify-between gap-3 py-3">
              <div className="flex flex-wrap items-center gap-2">
                <span className="font-medium text-slate-700">{s.standing}</span>
                {s.is_current && (
                  <Tag color="success" bordered={false}>
                    ปัจจุบัน
                  </Tag>
                )}
                <span className="text-xs text-slate-400">
                  มีผล <span className="num">{dateText(s.effective_date)}</span>
                </span>
              </div>
              <DeleteRowButton
                title="ลบวิทยฐานะนี้?"
                onConfirm={() =>
                  deleteMutation.mutate(s.id, {
                    onSuccess: () => message.success("ลบแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              />
            </li>
          ))}
        </ul>
      )}
      <AddBar>
        <StandingAddForm onAdd={handleAdd} submitting={createMutation.isPending} />
      </AddBar>
    </SectionCard>
  );
}

// PersonnelSubResources รวมตำแหน่งบริหาร + วิทยฐานะ (แสดงในหน้าแก้ไขบุคลากร)
export function PersonnelSubResources({ personnelId }: { personnelId: string }): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <PositionsSection personnelId={personnelId} />
      <StandingsSection personnelId={personnelId} />
    </div>
  );
}
