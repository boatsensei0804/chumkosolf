"use client";

import { ApartmentOutlined, DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined, TeamOutlined } from "@ant-design/icons";
import { App, Button, Drawer, Empty, Input, Popconfirm, Spin, Table, Tag, type TableProps } from "antd";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { useDeleteClass } from "@/features/classes/hooks";
import { useDirectoryClassStudents, useDirectoryClasses, useDirectoryStudentSearch } from "@/features/directory/hooks";
import { canManageTimetable } from "@/features/navigation/menu";
import type { DirectoryClass, DirectoryStudentClass } from "@/shared/schemas/directory";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

function fullName(p: { prefix: string; first_name: string; last_name: string }): string {
  return `${p.prefix}${p.first_name} ${p.last_name}`.trim();
}

// รายชื่อนักเรียนในห้อง (ข้อมูลพื้นฐาน) — โหลดเมื่อเปิด drawer
function ClassRoster({ classId }: { classId: string }): ReactNode {
  const { data, isLoading, isError, error } = useDirectoryClassStudents(classId);
  if (isLoading) {
    return (
      <div className="flex justify-center py-10">
        <Spin />
      </div>
    );
  }
  if (isError) {
    return <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลไม่สำเร็จ: {error?.message}</div>;
  }
  const students = data ?? [];
  if (students.length === 0) return <Empty description="ยังไม่มีนักเรียนในห้องนี้" />;
  return (
    <ol className="flex flex-col divide-y divide-slate-100">
      {students.map((s, i) => (
        <li key={s.student_id} className="flex items-center gap-3 py-2.5">
          <span className="num w-6 shrink-0 text-right text-xs text-slate-400">{i + 1}</span>
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-brand/10 text-xs font-semibold text-brand">
            {(s.first_name.charAt(0) || "?").toUpperCase()}
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate text-sm text-slate-800">{fullName(s)}</span>
            <span className="num block text-xs text-slate-400">{s.student_code}</span>
          </span>
        </li>
      ))}
    </ol>
  );
}

export default function ClassesPage(): ReactNode {
  const { message } = App.useApp();
  const router = useRouter();
  const { user } = useAuth();
  const canEdit = user ? canManageTimetable(user) : false; // วิชาการ/แอดมิน แก้ไขได้; ครูอื่นดูอย่างเดียว

  const [q, setQ] = useState("");
  const searching = q.trim().length > 0;
  const { data: classes, isLoading: classesLoading, isError: classesError } = useDirectoryClasses();
  const { data: results, isLoading: searchLoading } = useDirectoryStudentSearch(q.trim());
  const [openClass, setOpenClass] = useState<DirectoryClass | null>(null);
  const deleteMutation = useDeleteClass();

  const searchColumns: TableProps<DirectoryStudentClass>["columns"] = [
    {
      title: "ชื่อ-นามสกุล",
      key: "name",
      render: (_, r) => (
        <div className="leading-tight">
          <div className="text-sm font-medium text-slate-800">{fullName(r)}</div>
          <div className="num text-xs text-slate-400">{r.student_code}</div>
        </div>
      ),
    },
    {
      title: "อยู่ห้อง",
      dataIndex: "class_label",
      key: "class_label",
      render: (v: string) => (
        <Tag bordered={false} color="blue">
          {v || "—"}
        </Tag>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<ApartmentOutlined />}
        title="ห้องเรียน"
        subtitle={canEdit ? "จัดการห้องที่ปรึกษา ค้นหาและดูรายชื่อนักเรียน" : "ดูห้องเรียนและค้นหาว่านักเรียนอยู่ห้องไหน (ดูอย่างเดียว)"}
        actions={
          canEdit ? (
            <Link href="/classes/new">
              <Button type="primary" icon={<PlusOutlined />}>
                เพิ่มห้องเรียน
              </Button>
            </Link>
          ) : undefined
        }
      />

      <SectionCard icon={<SearchOutlined />} title="ค้นหานักเรียน">
        <Input
          allowClear
          size="large"
          prefix={<SearchOutlined className="text-slate-400" />}
          placeholder="พิมพ์ชื่อ นามสกุล หรือรหัสนักเรียน"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="ค้นหานักเรียน"
        />
        {searching && (
          <div className="mt-4">
            <Table<DirectoryStudentClass>
              rowKey="student_id"
              size="small"
              columns={searchColumns}
              dataSource={results ?? []}
              loading={searchLoading}
              pagination={false}
              scroll={{ x: 420 }}
              locale={{ emptyText: "ไม่พบนักเรียนที่ตรงกับคำค้น" }}
            />
          </div>
        )}
      </SectionCard>

      {!searching && (
        <SectionCard icon={<TeamOutlined />} title="ห้องเรียนทั้งหมด" accent="violet">
          {classesLoading ? (
            <div className="flex justify-center py-10">
              <Spin />
            </div>
          ) : classesError ? (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลห้องเรียนไม่สำเร็จ</div>
          ) : (classes ?? []).length === 0 ? (
            <Empty description="ยังไม่มีห้องเรียนในเทอมนี้" />
          ) : (
            <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-4">
              {(classes ?? []).map((c) => (
                <div
                  key={c.id}
                  className="flex flex-col items-start rounded-xl border border-slate-200 bg-white p-3 transition-colors hover:border-brand hover:bg-slate-50"
                >
                  <button
                    type="button"
                    onClick={() => {
                      if (c.student_count === 0) {
                        message.info("ห้องนี้ยังไม่มีนักเรียน");
                        return;
                      }
                      setOpenClass(c);
                    }}
                    className="flex w-full flex-col items-start text-left"
                  >
                    <span className="text-sm font-semibold text-slate-800">
                      {c.grade_level} {c.room_name}
                    </span>
                    <span className="num mt-1 text-xs text-slate-400">{c.student_count} คน</span>
                  </button>
                  {canEdit && (
                    <div className="mt-2 flex gap-1">
                      <Button type="text" size="small" icon={<EditOutlined />} aria-label="จัดการ" onClick={() => router.push(`/classes/${c.id}/edit`)} />
                      <Popconfirm
                        title="ยืนยันการลบ"
                        description={`ลบห้อง ${c.grade_level} ${c.room_name}?`}
                        okText="ลบ"
                        cancelText="ยกเลิก"
                        okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
                        onConfirm={() =>
                          deleteMutation.mutate(c.id, {
                            onSuccess: () => message.success("ลบห้องเรียนแล้ว"),
                            onError: (err) => message.error(err.message),
                          })
                        }
                      >
                        <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
                      </Popconfirm>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </SectionCard>
      )}

      <Drawer
        title={openClass ? `รายชื่อห้อง ${openClass.grade_level} ${openClass.room_name}` : "รายชื่อนักเรียน"}
        placement="right"
        width={420}
        open={openClass !== null}
        onClose={() => setOpenClass(null)}
        destroyOnHidden
      >
        {openClass && <ClassRoster classId={openClass.id} />}
      </Drawer>
    </div>
  );
}
