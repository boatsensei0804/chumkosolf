"use client";

import { DeleteOutlined, EditOutlined, PlusOutlined, UserOutlined } from "@ant-design/icons";
import { App, Button, Popconfirm, Table, type TableProps } from "antd";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { useDeleteStudent, useStudentList } from "@/features/students/hooks";
import { PageHeader } from "@/shared/ui/PageHeader";
import type { StudentListItem } from "@/shared/schemas/student";

const PAGE_SIZE = 20;

export default function StudentListPage(): ReactNode {
  const { message } = App.useApp();
  const router = useRouter();
  const [page, setPage] = useState(1);
  const { data, isLoading, isError, error } = useStudentList(page, PAGE_SIZE);
  const deleteMutation = useDeleteStudent();

  const columns: TableProps<StudentListItem>["columns"] = [
    {
      title: "ชื่อ-นามสกุล",
      key: "name",
      render: (_, r) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-sky-50 text-xs font-semibold text-sky-600">
            {(r.first_name.charAt(0) || "?").toUpperCase()}
          </span>
          <div className="leading-tight">
            <div className="font-medium text-slate-800">{`${r.prefix}${r.first_name} ${r.last_name}`.trim()}</div>
            <div className="num text-xs text-slate-400">{r.student_code}</div>
          </div>
        </div>
      ),
    },
    { title: "เลขบัตรประชาชน", dataIndex: "national_id_masked", key: "nid", render: (v: string) => <span className="num text-slate-500">{v}</span> },
    { title: "เบอร์โทร", dataIndex: "phone", key: "phone", render: (v: string) => <span className="num text-slate-500">{v || "—"}</span> },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 100,
      render: (_, r) => (
        <div className="flex justify-end gap-1">
          <Button type="text" size="small" icon={<EditOutlined />} aria-label="แก้ไข" onClick={() => router.push(`/students/${r.id}/edit`)} />
          <Popconfirm
            title="ยืนยันการลบ"
            description={`ลบ ${r.first_name} ${r.last_name}?`}
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
            onConfirm={() =>
              deleteMutation.mutate(r.id, {
                onSuccess: () => message.success("ลบข้อมูลนักเรียนแล้ว"),
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
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<UserOutlined />}
        title="นักเรียน"
        subtitle={`ข้อมูลนักเรียน${typeof data?.total === "number" ? ` · ${data.total} คน` : ""}`}
        actions={
          <Link href="/students/new">
            <Button type="primary" icon={<PlusOutlined />}>
              เพิ่มนักเรียน
            </Button>
          </Link>
        }
      />

      {isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลไม่สำเร็จ: {error?.message}</div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <Table<StudentListItem>
            rowKey="id"
            size="middle"
            columns={columns}
            dataSource={data?.items ?? []}
            loading={isLoading}
            scroll={{ x: 640 }}
            pagination={{ current: page, pageSize: PAGE_SIZE, total: data?.total ?? 0, onChange: setPage, showSizeChanger: false, className: "px-4" }}
            locale={{ emptyText: "ยังไม่มีข้อมูลนักเรียน — กด “เพิ่มนักเรียน” เพื่อเริ่ม" }}
          />
        </div>
      )}
    </div>
  );
}
