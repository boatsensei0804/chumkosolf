"use client";

import { DeleteOutlined, KeyOutlined, PlusOutlined } from "@ant-design/icons";
import { App, Button, Input, Popconfirm, Spin, Table, Tag, type TableProps } from "antd";
import { useState, type ReactNode } from "react";

import { useCreateKioskAccount, useDeleteKioskAccount, useKioskAccounts } from "@/features/kiosk-accounts/hooks";
import type { KioskAccount } from "@/shared/schemas/kioskAccount";
import { SectionCard } from "@/shared/ui/SectionCard";

// แผงจัดการบัญชีเครื่องสแกนหน้า (ใช้เป็นแท็บในหน้าสแกนหน้า — school admin เท่านั้น)
export function KioskAccountsPanel(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading, isError, error } = useKioskAccounts();
  const createMutation = useCreateKioskAccount();
  const deleteMutation = useDeleteKioskAccount();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const submit = (): void => {
    if (username.trim().length < 3 || password.length < 6) {
      message.error("ชื่อผู้ใช้อย่างน้อย 3 ตัว และรหัสผ่านอย่างน้อย 6 ตัว");
      return;
    }
    createMutation.mutate(
      { username: username.trim(), password },
      {
        onSuccess: () => {
          message.success("สร้างบัญชีสแกนหน้าแล้ว");
          setUsername("");
          setPassword("");
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  const columns: TableProps<KioskAccount>["columns"] = [
    { title: "ชื่อผู้ใช้", dataIndex: "username", key: "username", render: (v: string) => <span className="font-medium text-slate-800">{v}</span> },
    {
      title: "สถานะ",
      dataIndex: "is_active",
      key: "is_active",
      render: (a: boolean) => (a ? <Tag color="success" bordered={false}>ใช้งาน</Tag> : <Tag bordered={false}>ปิด</Tag>),
    },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 80,
      render: (_, r) => (
        <Popconfirm
          title="ลบบัญชีนี้?"
          okText="ลบ"
          cancelText="ยกเลิก"
          okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
          onConfirm={() =>
            deleteMutation.mutate(r.id, {
              onSuccess: () => message.success("ลบบัญชีแล้ว"),
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
    <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
      <SectionCard className="lg:col-span-2" icon={<KeyOutlined />} title="บัญชีทั้งหมด">
        {isError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดไม่สำเร็จ: {error?.message}</div>
        ) : (
          <Table<KioskAccount>
            rowKey="id"
            size="middle"
            columns={columns}
            dataSource={data ?? []}
            loading={isLoading}
            pagination={false}
            locale={{ emptyText: "ยังไม่มีบัญชีสแกนหน้า" }}
          />
        )}
      </SectionCard>

      <SectionCard icon={<PlusOutlined />} title="สร้างบัญชีใหม่" accent="emerald">
        <div className="flex flex-col gap-3">
          <div>
            <label className="mb-1 block text-sm text-slate-600">ชื่อผู้ใช้</label>
            <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="เช่น gate-1" aria-label="ชื่อผู้ใช้" />
          </div>
          <div>
            <label className="mb-1 block text-sm text-slate-600">รหัสผ่าน</label>
            <Input.Password value={password} onChange={(e) => setPassword(e.target.value)} placeholder="อย่างน้อย 6 ตัว" aria-label="รหัสผ่าน" />
          </div>
          <Button type="primary" icon={<PlusOutlined />} loading={createMutation.isPending} onClick={submit} block>
            สร้างบัญชี
          </Button>
          {isLoading && <Spin size="small" />}
          <p className="text-xs text-slate-400">นำชื่อผู้ใช้/รหัสนี้ไปล็อกอินบนเครื่อง kiosk แล้วเปิดหน้าสแกนหน้า</p>
        </div>
      </SectionCard>
    </div>
  );
}
