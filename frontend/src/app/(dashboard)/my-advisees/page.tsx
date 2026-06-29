"use client";

import { EditOutlined, TeamOutlined } from "@ant-design/icons";
import { App, Button, Drawer, Spin, Table, Tag, type TableProps } from "antd";
import { useState, type ReactNode } from "react";

import { AdviseeEditForm } from "@/features/me/AdviseeEditForm";
import { toAdviseeFormValues, toUpdateMyAdviseeBody, type AdviseeFormValues } from "@/features/me/adviseeForm";
import { useMyAdvisee, useMyAdvisees, useUpdateMyAdvisee } from "@/features/me/hooks";
import { attendanceStatusLabel, type AttendanceStatus } from "@/shared/schemas/enums";
import type { Advisee } from "@/shared/schemas/me";
import { PageHeader } from "@/shared/ui/PageHeader";

const STATUS_COLOR: Record<AttendanceStatus, string> = {
  present: "success",
  late: "warning",
  absent: "error",
  sick_leave: "blue",
  personal_leave: "purple",
};

function StatusTag({ status }: { status: Advisee["today_status"] }): ReactNode {
  if (status === "") {
    return (
      <Tag bordered={false} color="default">
        ยังไม่เช็ค
      </Tag>
    );
  }
  return (
    <Tag bordered={false} color={STATUS_COLOR[status]}>
      {attendanceStatusLabel[status]}
    </Tag>
  );
}

// เนื้อหา Drawer แก้ไข — โหลดข้อมูลนักเรียนของห้องที่ปรึกษา แล้วแสดงฟอร์ม
function AdviseeEditBody({ studentId, onSaved }: { studentId: string; onSaved: () => void }): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading, isError, error } = useMyAdvisee(studentId);
  const updateMutation = useUpdateMyAdvisee(studentId);
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: AdviseeFormValues): void => {
    setErrorMessage("");
    updateMutation.mutate(toUpdateMyAdviseeBody(values), {
      onSuccess: () => {
        message.success("บันทึกข้อมูลนักเรียนแล้ว");
        onSaved();
      },
      onError: (err) => setErrorMessage(err.message),
    });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-10">
        <Spin />
      </div>
    );
  }
  if (isError || !data) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
        โหลดข้อมูลไม่สำเร็จ: {error?.message ?? "ไม่พบข้อมูลนักเรียน"}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-500">
        รหัสนักเรียน <span className="num text-slate-700">{data.student_code}</span> · เลขบัตร{" "}
        <span className="num text-slate-700">{data.national_id_masked || "—"}</span>
        <div className="mt-0.5">รหัสนักเรียน/สถานะ/เลขบัตร แก้ไขได้ที่กลุ่มวิชาการเท่านั้น</div>
      </div>
      <AdviseeEditForm
        defaultValues={toAdviseeFormValues(data)}
        onSubmit={handleSubmit}
        isSubmitting={updateMutation.isPending}
        errorMessage={errorMessage}
      />
    </div>
  );
}

export default function MyAdviseesPage(): ReactNode {
  const { data, isLoading, isError, error } = useMyAdvisees();
  const advisees = data ?? [];
  const [editing, setEditing] = useState<Advisee | null>(null);

  const columns: TableProps<Advisee>["columns"] = [
    {
      title: "ชื่อ-นามสกุล",
      key: "name",
      render: (_, r) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand/10 text-xs font-semibold text-brand">
            {(r.first_name.charAt(0) || "?").toUpperCase()}
          </span>
          <div className="leading-tight">
            <div className="font-medium text-slate-800">
              {`${r.prefix}${r.first_name} ${r.last_name}`.trim()}
            </div>
            <div className="num text-xs text-slate-400">{r.student_code}</div>
          </div>
        </div>
      ),
    },
    {
      title: "ห้อง",
      dataIndex: "class_label",
      key: "class_label",
      render: (v: string) => <span className="text-slate-600">{v || "—"}</span>,
    },
    {
      title: "เลขบัตรประชาชน",
      dataIndex: "national_id_masked",
      key: "national_id",
      render: (v: string) => <span className="num text-slate-500">{v || "—"}</span>,
    },
    {
      title: "เบอร์โทร",
      dataIndex: "phone",
      key: "phone",
      render: (v: string) => <span className="num text-slate-500">{v || "—"}</span>,
    },
    {
      title: "เช็คชื่อวันนี้",
      dataIndex: "today_status",
      key: "today_status",
      render: (_, r) => <StatusTag status={r.today_status} />,
    },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 100,
      render: (_, r) => (
        <Button type="text" size="small" icon={<EditOutlined />} aria-label="แก้ไข" onClick={() => setEditing(r)}>
          แก้ไข
        </Button>
      ),
    },
  ];

  const editingName = editing ? `${editing.prefix}${editing.first_name} ${editing.last_name}`.trim() : "";

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<TeamOutlined />}
        title="นักเรียนที่ปรึกษาของฉัน"
        subtitle={`รายชื่อนักเรียนในห้องที่ปรึกษาของคุณ${advisees.length > 0 ? ` · ${advisees.length} คน` : ""}`}
      />

      {isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          โหลดข้อมูลไม่สำเร็จ: {error?.message}
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <Table<Advisee>
            rowKey="student_id"
            size="middle"
            columns={columns}
            dataSource={advisees}
            loading={isLoading}
            scroll={{ x: 760 }}
            pagination={false}
            locale={{
              emptyText: "คุณยังไม่ได้เป็นครูที่ปรึกษาของห้องใดในเทอมนี้",
            }}
          />
        </div>
      )}

      <Drawer
        title={editingName ? `แก้ไขข้อมูล · ${editingName}` : "แก้ไขข้อมูลนักเรียน"}
        placement="right"
        width={560}
        open={editing !== null}
        onClose={() => setEditing(null)}
        destroyOnHidden
      >
        {editing && <AdviseeEditBody studentId={editing.student_id} onSaved={() => setEditing(null)} />}
      </Drawer>
    </div>
  );
}
