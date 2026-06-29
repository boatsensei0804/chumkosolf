"use client";

import { DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined } from "@ant-design/icons";
import { App, Button, Input, Popconfirm, Table, Tag, type TableProps } from "antd";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";

import { useDeleteStudent, useStudentList } from "@/features/students/hooks";
import { isStudentStatus, studentStatusColor, studentStatusLabel, type StudentListItem } from "@/shared/schemas/student";

const PAGE_SIZE = 20;

// แผงรายชื่อนักเรียน + ค้นหา (ใช้ทั้งหน้านักเรียนเดี่ยว และเป็นแท็บในหน้าข้อมูลบุคคล)
export function StudentsListPanel(): ReactNode {
  const { message } = App.useApp();
  const router = useRouter();
  const [page, setPage] = useState(1);
  const [input, setInput] = useState("");
  const [search, setSearch] = useState("");
  const { data, isLoading, isError, error } = useStudentList(page, PAGE_SIZE, search);
  const deleteMutation = useDeleteStudent();

  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(input.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [input]);

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
    {
      title: "สถานะ",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (v: string) =>
        isStudentStatus(v) ? (
          <Tag color={studentStatusColor[v]} bordered={false}>
            {studentStatusLabel[v]}
          </Tag>
        ) : (
          <span className="text-slate-400">—</span>
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
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Input
          allowClear
          prefix={<SearchOutlined className="text-slate-400" />}
          placeholder="ค้นหาชื่อ นามสกุล หรือรหัสนักเรียน"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          className="max-w-xs"
          aria-label="ค้นหานักเรียน"
        />
        <Link href="/students/new">
          <Button type="primary" icon={<PlusOutlined />}>
            เพิ่มนักเรียน
          </Button>
        </Link>
      </div>

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
            locale={{ emptyText: search ? "ไม่พบนักเรียนที่ค้นหา" : "ยังไม่มีข้อมูลนักเรียน — กด “เพิ่มนักเรียน” เพื่อเริ่ม" }}
          />
        </div>
      )}
    </div>
  );
}
