"use client";

import { DeleteOutlined } from "@ant-design/icons";
import { App, Button, Popconfirm, Select, Table, Tag, type TableProps } from "antd";
import { useState, type ReactNode } from "react";

import { useClassList } from "@/features/classes/hooks";
import { usePersonnelList } from "@/features/personnel/hooks";
import type { AssignmentBody, TeachingAssignment } from "@/shared/schemas/timetable";

import { useAssignments, useCreateAssignment, useDeleteAssignment, useSubjects } from "./hooks";

export function AssignmentsTab(): ReactNode {
  const { message } = App.useApp();
  const { data: assignments, isLoading } = useAssignments();
  const { data: subjects } = useSubjects();
  const { data: classes } = useClassList();
  const { data: personnel } = usePersonnelList(1, 200);
  const createMutation = useCreateAssignment();
  const deleteMutation = useDeleteAssignment();

  const [personnelId, setPersonnelId] = useState("");
  const [subjectId, setSubjectId] = useState("");
  const [classId, setClassId] = useState("");

  const handleAdd = (): void => {
    if (personnelId === "" || subjectId === "" || classId === "") {
      message.warning("กรุณาเลือกครู วิชา และห้องให้ครบ");
      return;
    }
    const body: AssignmentBody = { personnel_id: personnelId, subject_id: subjectId, class_id: classId };
    createMutation.mutate(body, {
      onSuccess: () => {
        message.success("เพิ่มการมอบหมายแล้ว");
        setSubjectId("");
      },
      onError: (err) => message.error(err.message),
    });
  };

  const columns: TableProps<TeachingAssignment>["columns"] = [
    {
      title: "ห้อง",
      key: "class",
      width: 110,
      render: (_, r) => (
        <Tag color="violet" bordered={false}>
          {r.grade_level} {r.room_name}
        </Tag>
      ),
    },
    {
      title: "วิชา",
      key: "subject",
      render: (_, r) => (
        <div>
          <span className="font-medium text-slate-800">{r.subject_name}</span>{" "}
          <span className="num text-xs text-slate-400">{r.subject_code}</span>
        </div>
      ),
    },
    { title: "ครูผู้สอน", dataIndex: "teacher_name", key: "teacher", render: (v: string) => <span className="text-slate-600">{v}</span> },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 60,
      render: (_, r) => (
        <Popconfirm
          title="ลบการมอบหมายนี้?"
          okText="ลบ"
          cancelText="ยกเลิก"
          okButtonProps={{ danger: true }}
          onConfirm={() =>
            deleteMutation.mutate(r.id, {
              onSuccess: () => message.success("ลบแล้ว"),
              onError: (err) => message.error(err.message),
            })
          }
        >
          <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <Table<TeachingAssignment>
          rowKey="id"
          size="middle"
          columns={columns}
          dataSource={assignments ?? []}
          loading={isLoading}
          scroll={{ x: 520 }}
          pagination={false}
          locale={{ emptyText: "ยังไม่มีการมอบหมายการสอนในเทอมนี้" }}
        />
      </div>

      <div className="flex flex-wrap items-end gap-3 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ครูผู้สอน</label>
          <Select
            showSearch
            value={personnelId || undefined}
            onChange={setPersonnelId}
            placeholder="เลือกครู"
            style={{ width: 200 }}
            optionFilterProp="label"
            options={(personnel?.items ?? []).map((p) => ({
              value: p.id,
              label: `${p.prefix}${p.first_name} ${p.last_name}`,
            }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วิชา</label>
          <Select
            showSearch
            value={subjectId || undefined}
            onChange={setSubjectId}
            placeholder="เลือกวิชา"
            style={{ width: 220 }}
            optionFilterProp="label"
            options={(subjects ?? []).map((s) => ({ value: s.id, label: `${s.subject_code} ${s.name}` }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">ห้อง</label>
          <Select
            showSearch
            value={classId || undefined}
            onChange={setClassId}
            placeholder="เลือกห้อง"
            style={{ width: 150 }}
            optionFilterProp="label"
            options={(classes ?? []).map((c) => ({ value: c.id, label: `${c.grade_level} ${c.room_name}` }))}
          />
        </div>
        <Button type="primary" loading={createMutation.isPending} onClick={handleAdd}>
          มอบหมาย
        </Button>
      </div>
    </div>
  );
}
