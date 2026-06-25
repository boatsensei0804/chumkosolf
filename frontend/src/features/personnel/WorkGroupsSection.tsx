"use client";

import { ApartmentOutlined, DeleteOutlined } from "@ant-design/icons";
import { App, Button, Checkbox, Popconfirm, Select, Spin, Tag } from "antd";
import { useState, type ReactNode } from "react";

import { SectionCard } from "@/shared/ui/SectionCard";
import { workGroupLabel } from "@/shared/schemas/enums";
import type { AssignWorkGroupBody, WorkGroup } from "@/shared/schemas/workGroup";

import {
  useAssignWorkGroup,
  usePersonnelWorkGroups,
  useUnassignWorkGroup,
  useWorkGroups,
} from "./workGroupHooks";

function groupLabel(code: string, name: string): string {
  return (workGroupLabel as Record<string, string>)[code] ?? name;
}

// ===== ฟอร์มมอบหมายกลุ่มงาน (presentational) =====
export function WorkGroupAddForm(props: {
  groups: WorkGroup[];
  onAdd: (body: AssignWorkGroupBody) => void;
  submitting: boolean;
}): ReactNode {
  const { groups, onAdd, submitting } = props;
  const [selectedId, setSelectedId] = useState("");
  const [isGroupAdmin, setIsGroupAdmin] = useState(false);

  if (groups.length === 0) {
    return <p className="text-sm text-slate-400">มอบหมายครบทุกกลุ่มงานแล้ว</p>;
  }

  const effectiveId = groups.some((g) => g.id === selectedId) ? selectedId : (groups[0]?.id ?? "");

  return (
    <div className="flex flex-wrap items-center gap-3">
      <Select
        value={effectiveId}
        onChange={setSelectedId}
        style={{ width: 220 }}
        aria-label="เลือกกลุ่มงาน"
        options={groups.map((g) => ({ value: g.id, label: groupLabel(g.code, g.name) }))}
      />
      <Checkbox checked={isGroupAdmin} onChange={(e) => setIsGroupAdmin(e.target.checked)}>
        เป็นหัวหน้ากลุ่ม
      </Checkbox>
      <Button
        type="primary"
        loading={submitting}
        onClick={() => onAdd({ work_group_id: effectiveId, is_group_admin: isGroupAdmin })}
      >
        มอบหมาย
      </Button>
    </div>
  );
}

// ===== section กลุ่มงานที่สังกัด (container) =====
export function WorkGroupsSection({ personnelId }: { personnelId: string }): ReactNode {
  const { message } = App.useApp();
  const { data: allGroups } = useWorkGroups();
  const { data: assigned, isLoading } = usePersonnelWorkGroups(personnelId);
  const assignMutation = useAssignWorkGroup(personnelId);
  const unassignMutation = useUnassignWorkGroup(personnelId);

  const assignedIds = new Set((assigned ?? []).map((m) => m.work_group_id));
  const available = (allGroups ?? []).filter((g) => !assignedIds.has(g.id));

  const handleAdd = (body: AssignWorkGroupBody): void => {
    assignMutation.mutate(body, {
      onSuccess: () => message.success("มอบหมายกลุ่มงานแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  return (
    <SectionCard
      icon={<ApartmentOutlined />}
      title="กลุ่มงานที่สังกัด"
      description="กำหนดสิทธิ์การเข้าถึงเมนูตามกลุ่มงาน"
    >
      {isLoading ? (
        <Spin />
      ) : (assigned?.length ?? 0) === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-8 text-center text-sm text-slate-400">
          ยังไม่ได้สังกัดกลุ่มงาน
        </div>
      ) : (
        <ul className="divide-y divide-slate-100">
          {(assigned ?? []).map((m) => (
            <li key={m.work_group_id} className="flex items-center justify-between gap-3 py-3">
              <div className="flex flex-wrap items-center gap-2">
                <Tag color="blue" bordered={false}>
                  {groupLabel(m.code, m.name)}
                </Tag>
                {m.is_group_admin && (
                  <Tag color="gold" bordered={false}>
                    หัวหน้ากลุ่ม
                  </Tag>
                )}
              </div>
              <Popconfirm
                title="ถอดออกจากกลุ่มงานนี้?"
                okText="ถอด"
                cancelText="ยกเลิก"
                okButtonProps={{ danger: true }}
                onConfirm={() =>
                  unassignMutation.mutate(m.work_group_id, {
                    onSuccess: () => message.success("ถอดออกแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              >
                <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ถอด" />
              </Popconfirm>
            </li>
          ))}
        </ul>
      )}
      <div className="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
        <WorkGroupAddForm groups={available} onAdd={handleAdd} submitting={assignMutation.isPending} />
      </div>
    </SectionCard>
  );
}
