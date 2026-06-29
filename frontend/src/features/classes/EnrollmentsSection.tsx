"use client";

import { DeleteOutlined, TeamOutlined } from "@ant-design/icons";
import { App, Button, Empty, Popconfirm, Select, Spin } from "antd";
import { useState, type ReactNode } from "react";

import { useStudentList } from "@/features/students/hooks";
import type { StudentListItem } from "@/shared/schemas/student";
import { SectionCard } from "@/shared/ui/SectionCard";

import { useEnrollStudentsBulk, useEnrollments, useRemoveEnrollment } from "./hooks";

// ฟอร์มจัดนักเรียนเข้าห้องแบบเลือกหลายคนพร้อมกัน (presentational)
export function EnrollAddForm(props: {
  students: StudentListItem[];
  onAdd: (studentIds: string[]) => void;
  submitting: boolean;
}): ReactNode {
  const { students, onAdd, submitting } = props;
  const [ids, setIds] = useState<string[]>([]);

  if (students.length === 0) {
    return <p className="text-sm text-slate-400">จัดนักเรียนครบแล้ว หรือยังไม่มีนักเรียนในระบบ</p>;
  }

  const submit = (): void => {
    if (ids.length === 0) return;
    onAdd(ids);
    setIds([]);
  };

  return (
    <div className="flex flex-col gap-3">
      <div>
        <label className="mb-1 block text-xs text-slate-500">เลือกนักเรียน (เลือกได้หลายคน)</label>
        <Select
          mode="multiple"
          showSearch
          optionFilterProp="label"
          value={ids}
          onChange={setIds}
          placeholder="พิมพ์ชื่อหรือรหัสเพื่อค้นหา แล้วเลือกได้หลายคน"
          className="w-full"
          maxTagCount="responsive"
          options={students.map((s) => ({
            value: s.id,
            label: `${s.student_code} · ${s.first_name} ${s.last_name}`.trim(),
          }))}
        />
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Button type="primary" loading={submitting} disabled={ids.length === 0} onClick={submit}>
          จัดเข้าห้อง{ids.length > 0 ? ` (${ids.length})` : ""}
        </Button>
        <Button type="link" size="small" disabled={ids.length === students.length} onClick={() => setIds(students.map((s) => s.id))}>
          เลือกทั้งหมด ({students.length})
        </Button>
        {ids.length > 0 && (
          <Button type="text" size="small" onClick={() => setIds([])}>
            ล้าง
          </Button>
        )}
      </div>
    </div>
  );
}

export function EnrollmentsSection({ classId }: { classId: string }): ReactNode {
  const { message } = App.useApp();
  const { data: enrollments, isLoading } = useEnrollments(classId);
  const { data: students } = useStudentList(1, 100);
  const bulkMutation = useEnrollStudentsBulk(classId);
  const removeMutation = useRemoveEnrollment(classId);

  const enrolledIds = new Set((enrollments ?? []).map((e) => e.student_id));
  const available = (students?.items ?? []).filter((s) => !enrolledIds.has(s.id));

  return (
    <SectionCard icon={<TeamOutlined />} title="นักเรียนในห้อง" description="จัดนักเรียนเข้าห้อง (1 คนได้ห้องเดียวต่อเทอม)" accent="violet">
      {isLoading ? (
        <Spin />
      ) : (enrollments?.length ?? 0) === 0 ? (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="ยังไม่มีนักเรียนในห้อง" />
      ) : (
        <ul className="divide-y divide-slate-100">
          {(enrollments ?? []).map((e) => (
            <li key={e.id} className="flex items-center justify-between gap-3 py-2.5">
              <div className="flex items-center gap-2">
                <span className="num w-7 text-center text-xs text-slate-400">{e.student_no ?? "—"}</span>
                <span className="font-medium text-slate-700">
                  {e.prefix}
                  {e.first_name} {e.last_name}
                </span>
                <span className="num text-xs text-slate-400">{e.student_code}</span>
              </div>
              <Popconfirm
                title="ถอนนักเรียนคนนี้?"
                okText="ถอน"
                cancelText="ยกเลิก"
                okButtonProps={{ danger: true }}
                onConfirm={() =>
                  removeMutation.mutate(e.id, {
                    onSuccess: () => message.success("ถอนแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              >
                <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ถอน" />
              </Popconfirm>
            </li>
          ))}
        </ul>
      )}
      <div className="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
        <EnrollAddForm
          students={available}
          submitting={bulkMutation.isPending}
          onAdd={(ids) =>
            bulkMutation.mutate(ids, {
              onSuccess: () => message.success(`จัดนักเรียนเข้าห้องแล้ว ${ids.length} คน`),
              onError: (err) => message.error(err.message),
            })
          }
        />
      </div>
    </SectionCard>
  );
}
