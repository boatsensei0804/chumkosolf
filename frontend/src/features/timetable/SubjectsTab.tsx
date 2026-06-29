"use client";

import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import {
  App,
  Button,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Table,
  type TableProps,
} from "antd";
import { useState, type ReactNode } from "react";

import type { Subject, SubjectBody } from "@/shared/schemas/subject";

import { useCreateSubject, useDeleteSubject, useSubjects, useUpdateSubject } from "./hooks";

const emptySubject: SubjectBody = { subject_code: "", name: "", credit: null };

// ===== ฟอร์มกรอกข้อมูลรายวิชา (ใช้ทั้งเพิ่ม/แก้) =====
function SubjectFields(props: {
  value: SubjectBody;
  onChange: (v: SubjectBody) => void;
  error?: string;
}): ReactNode {
  const { value, onChange, error } = props;
  return (
    <div className="flex flex-wrap items-end gap-3">
      <div>
        <label className="mb-1 block text-xs text-slate-500">รหัสวิชา</label>
        <Input
          value={value.subject_code}
          onChange={(e) => onChange({ ...value, subject_code: e.target.value })}
          placeholder="เช่น ค21101"
          style={{ width: 140 }}
          status={error ? "error" : ""}
        />
      </div>
      <div className="min-w-[200px] flex-1">
        <label className="mb-1 block text-xs text-slate-500">ชื่อวิชา</label>
        <Input
          value={value.name}
          onChange={(e) => onChange({ ...value, name: e.target.value })}
          placeholder="เช่น คณิตศาสตร์พื้นฐาน"
          status={error ? "error" : ""}
        />
      </div>
      <div>
        <label className="mb-1 block text-xs text-slate-500">หน่วยกิต</label>
        <InputNumber
          min={0}
          max={99}
          step={0.5}
          value={value.credit}
          onChange={(v) => onChange({ ...value, credit: v ?? null })}
          style={{ width: 90 }}
        />
      </div>
    </div>
  );
}

// ===== ฟอร์มเพิ่มรายวิชา (presentational) =====
export function SubjectAddForm(props: { onAdd: (body: SubjectBody) => void; submitting: boolean }): ReactNode {
  const [value, setValue] = useState<SubjectBody>(emptySubject);
  const [error, setError] = useState("");

  const submit = (): void => {
    if (value.subject_code.trim() === "" || value.name.trim() === "") {
      setError("กรุณากรอกรหัสวิชาและชื่อวิชา");
      return;
    }
    setError("");
    props.onAdd({ ...value, subject_code: value.subject_code.trim(), name: value.name.trim() });
    setValue(emptySubject);
  };

  return (
    <div className="mt-4 flex flex-col gap-2 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
      <div className="flex flex-wrap items-end gap-3">
        <div className="flex-1">
          <SubjectFields value={value} onChange={setValue} error={error} />
        </div>
        <Button type="primary" loading={props.submitting} onClick={submit}>
          เพิ่มวิชา
        </Button>
      </div>
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
}

export function SubjectsTab(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useSubjects();
  const createMutation = useCreateSubject();
  const deleteMutation = useDeleteSubject();
  const [editing, setEditing] = useState<Subject | null>(null);
  const [editValue, setEditValue] = useState<SubjectBody>(emptySubject);
  const [editError, setEditError] = useState("");

  // mutation แก้ไขผูกกับ id ที่กำลังแก้
  const updateForEditing = useUpdateSubject(editing?.id ?? "");

  const handleAdd = (body: SubjectBody): void => {
    createMutation.mutate(body, {
      onSuccess: () => message.success("เพิ่มวิชาแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  const openEdit = (s: Subject): void => {
    setEditing(s);
    setEditValue({ subject_code: s.subject_code, name: s.name, credit: s.credit });
    setEditError("");
  };

  const saveEdit = (): void => {
    if (editValue.subject_code.trim() === "" || editValue.name.trim() === "") {
      setEditError("กรุณากรอกรหัสวิชาและชื่อวิชา");
      return;
    }
    updateForEditing.mutate(
      { ...editValue, subject_code: editValue.subject_code.trim(), name: editValue.name.trim() },
      {
        onSuccess: () => {
          message.success("บันทึกรายวิชาแล้ว");
          setEditing(null);
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  const columns: TableProps<Subject>["columns"] = [
    { title: "รหัสวิชา", dataIndex: "subject_code", key: "code", width: 130, render: (v: string) => <span className="num text-slate-600">{v}</span> },
    { title: "ชื่อวิชา", dataIndex: "name", key: "name", render: (v: string) => <span className="font-medium text-slate-800">{v}</span> },
    { title: "หน่วยกิต", dataIndex: "credit", key: "credit", width: 100, render: (v: number | null) => <span className="num text-slate-500">{v ?? "—"}</span> },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 90,
      render: (_, r) => (
        <div className="flex justify-end gap-1">
          <Button type="text" size="small" icon={<EditOutlined />} aria-label="แก้ไข" onClick={() => openEdit(r)} />
          <Popconfirm
            title="ลบวิชานี้?"
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true }}
            onConfirm={() =>
              deleteMutation.mutate(r.id, {
                onSuccess: () => message.success("ลบวิชาแล้ว"),
                onError: (err) => message.error(err.message),
              })
            }
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
          </Popconfirm>
        </div>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <Table<Subject>
          rowKey="id"
          size="middle"
          columns={columns}
          dataSource={data ?? []}
          loading={isLoading}
          scroll={{ x: 480 }}
          pagination={false}
          locale={{ emptyText: "ยังไม่มีรายวิชา — เพิ่มด้านล่าง" }}
        />
      </div>

      <SubjectAddForm onAdd={handleAdd} submitting={createMutation.isPending} />

      <Modal
        open={editing !== null}
        title="แก้ไขรายวิชา"
        onCancel={() => setEditing(null)}
        onOk={saveEdit}
        okText="บันทึก"
        cancelText="ยกเลิก"
        confirmLoading={updateForEditing.isPending}
        destroyOnHidden
      >
        <SubjectFields value={editValue} onChange={setEditValue} error={editError} />
        {editError && <p className="mt-2 text-sm text-red-500">{editError}</p>}
      </Modal>
    </div>
  );
}
