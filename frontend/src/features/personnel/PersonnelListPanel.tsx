"use client";

import { DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined } from "@ant-design/icons";
import { App, Button, Input, Popconfirm, Table, Tag, type TableProps } from "antd";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";

import { useDeletePersonnel, usePersonnelList } from "@/features/personnel/hooks";
import { personnelRoleLabel, type PersonnelListItem } from "@/shared/schemas/personnel";

const PAGE_SIZE = 20;

function roleText(role: string): string {
  if (role === "teacher" || role === "executive") return personnelRoleLabel[role];
  return role;
}

// แผงรายชื่อบุคลากร + ค้นหา (ใช้ทั้งหน้าบุคลากรเดี่ยว และเป็นแท็บในหน้าข้อมูลบุคคล)
export function PersonnelListPanel(): ReactNode {
  const { message } = App.useApp();
  const router = useRouter();
  const [page, setPage] = useState(1);
  const [input, setInput] = useState("");
  const [search, setSearch] = useState("");
  const { data, isLoading, isError, error } = usePersonnelList(page, PAGE_SIZE, search);
  const deleteMutation = useDeletePersonnel();

  // debounce การค้นหา + กลับไปหน้าแรกเมื่อคำค้นเปลี่ยน
  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(input.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [input]);

  const handleDelete = (id: string): void => {
    deleteMutation.mutate(id, {
      onSuccess: () => message.success("ลบข้อมูลบุคลากรแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  const columns: TableProps<PersonnelListItem>["columns"] = [
    {
      title: "ชื่อ-นามสกุล",
      key: "name",
      render: (_, r) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand/10 text-xs font-semibold text-brand">
            {(r.first_name.charAt(0) || "?").toUpperCase()}
          </span>
          <div className="leading-tight">
            <div className="font-medium text-slate-800">{`${r.prefix}${r.first_name} ${r.last_name}`.trim()}</div>
            <div className="text-xs text-slate-400">@{r.username}</div>
          </div>
        </div>
      ),
    },
    {
      title: "ตำแหน่ง",
      dataIndex: "role",
      key: "role",
      render: (v: string) => (
        <Tag color={v === "executive" ? "purple" : "blue"} bordered={false}>
          {roleText(v)}
        </Tag>
      ),
    },
    { title: "เลขบัตรประชาชน", dataIndex: "national_id_masked", key: "national_id", render: (v: string) => <span className="num text-slate-500">{v}</span> },
    { title: "เบอร์โทร", dataIndex: "phone", key: "phone", render: (v: string) => <span className="num text-slate-500">{v || "—"}</span> },
    {
      title: "สถานะ",
      dataIndex: "is_active",
      key: "is_active",
      render: (active: boolean) => (active ? <Tag color="success" bordered={false}>ใช้งาน</Tag> : <Tag bordered={false}>ระงับ</Tag>),
    },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 110,
      render: (_, r) => (
        <div className="flex justify-end gap-1">
          <Button type="text" size="small" icon={<EditOutlined />} aria-label="แก้ไข" onClick={() => router.push(`/personnel/${r.id}/edit`)} />
          <Popconfirm
            title="ยืนยันการลบ"
            description={`ลบ ${r.first_name} ${r.last_name}?`}
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
            onConfirm={() => handleDelete(r.id)}
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
          placeholder="ค้นหาชื่อ นามสกุล หรือเบอร์โทร"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          className="max-w-xs"
          aria-label="ค้นหาบุคลากร"
        />
        <Link href="/personnel/new">
          <Button type="primary" icon={<PlusOutlined />}>
            เพิ่มบุคลากร
          </Button>
        </Link>
      </div>

      {isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลไม่สำเร็จ: {error?.message}</div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <Table<PersonnelListItem>
            rowKey="id"
            size="middle"
            columns={columns}
            dataSource={data?.items ?? []}
            loading={isLoading}
            scroll={{ x: 760 }}
            pagination={{ current: page, pageSize: PAGE_SIZE, total: data?.total ?? 0, onChange: setPage, showSizeChanger: false, className: "px-4" }}
            locale={{ emptyText: search ? "ไม่พบบุคลากรที่ค้นหา" : "ยังไม่มีข้อมูลบุคลากร — กด “เพิ่มบุคลากร” เพื่อเริ่ม" }}
          />
        </div>
      )}
    </div>
  );
}
