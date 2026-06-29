"use client";

import { DeleteOutlined, SolutionOutlined } from "@ant-design/icons";
import { App, Button, Empty, Popconfirm, Select, Spin } from "antd";
import { useState, type ReactNode } from "react";

import { usePersonnelList } from "@/features/personnel/hooks";
import type { PersonnelListItem } from "@/shared/schemas/personnel";
import { SectionCard } from "@/shared/ui/SectionCard";

import { useAddAdvisor, useAdvisors, useRemoveAdvisor } from "./hooks";

// ฟอร์มเพิ่มครูที่ปรึกษา (presentational)
export function AdvisorAddForm(props: {
  personnel: PersonnelListItem[];
  onAdd: (personnelId: string) => void;
  submitting: boolean;
}): ReactNode {
  const { personnel, onAdd, submitting } = props;
  const [pid, setPid] = useState("");
  const [error, setError] = useState("");

  if (personnel.length === 0) {
    return <p className="text-sm text-slate-400">เพิ่มครูที่ปรึกษาครบทุกคนแล้ว หรือยังไม่มีบุคลากรในระบบ</p>;
  }

  const submit = (): void => {
    if (pid === "") {
      setError("กรุณาเลือกครู");
      return;
    }
    setError("");
    onAdd(pid);
    setPid("");
  };

  return (
    <div className="flex flex-wrap items-center gap-3">
      <Select
        showSearch
        optionFilterProp="label"
        value={pid || undefined}
        onChange={setPid}
        placeholder="เลือกครู"
        style={{ width: 240 }}
        status={error ? "error" : ""}
        options={personnel.map((p) => ({ value: p.id, label: `${p.prefix}${p.first_name} ${p.last_name}`.trim() }))}
      />
      <Button type="primary" loading={submitting} onClick={submit}>
        เพิ่มครูที่ปรึกษา
      </Button>
    </div>
  );
}

export function AdvisorsSection({ classId }: { classId: string }): ReactNode {
  const { message } = App.useApp();
  const { data: advisors, isLoading } = useAdvisors(classId);
  const { data: personnel } = usePersonnelList(1, 100);
  const addMutation = useAddAdvisor(classId);
  const removeMutation = useRemoveAdvisor(classId);

  const usedIds = new Set((advisors ?? []).map((a) => a.personnel_id));
  const available = (personnel?.items ?? []).filter((p) => !usedIds.has(p.id));

  return (
    <SectionCard icon={<SolutionOutlined />} title="ครูที่ปรึกษา" description="มอบหมายครูที่ปรึกษาประจำห้อง (มีได้หลายคน)" accent="amber">
      {isLoading ? (
        <Spin />
      ) : (advisors?.length ?? 0) === 0 ? (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="ยังไม่มีครูที่ปรึกษา" />
      ) : (
        <ul className="divide-y divide-slate-100">
          {(advisors ?? []).map((a) => (
            <li key={a.id} className="flex items-center justify-between gap-3 py-2.5">
              <span className="font-medium text-slate-700">
                {a.prefix}
                {a.first_name} {a.last_name}
              </span>
              <Popconfirm
                title="ถอดครูที่ปรึกษานี้?"
                okText="ถอด"
                cancelText="ยกเลิก"
                okButtonProps={{ danger: true }}
                onConfirm={() =>
                  removeMutation.mutate(a.id, {
                    onSuccess: () => message.success("ถอดแล้ว"),
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
        <AdvisorAddForm
          personnel={available}
          submitting={addMutation.isPending}
          onAdd={(personnelId) =>
            addMutation.mutate(
              { personnel_id: personnelId },
              { onSuccess: () => message.success("เพิ่มครูที่ปรึกษาแล้ว"), onError: (err) => message.error(err.message) },
            )
          }
        />
      </div>
    </SectionCard>
  );
}
