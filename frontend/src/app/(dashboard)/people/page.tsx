"use client";

import { IdcardOutlined, TeamOutlined, UserOutlined } from "@ant-design/icons";
import { Tabs, type TabsProps } from "antd";
import type { ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { isSchoolAdmin } from "@/features/navigation/menu";
import { PersonnelListPanel } from "@/features/personnel/PersonnelListPanel";
import { StudentsListPanel } from "@/features/students/StudentsListPanel";
import { PageHeader } from "@/shared/ui/PageHeader";

// หน้า "ข้อมูลบุคคล" รวมบุคลากร + นักเรียน/ผู้ปกครอง เป็นแท็บ — แสดงเฉพาะแท็บที่ผู้ใช้มีสิทธิ์
export default function PeoplePage(): ReactNode {
  const { user } = useAuth();
  if (!user) return null;

  const admin = isSchoolAdmin(user);
  const canPersonnel = admin || user.work_groups.some((g) => g.code === "personnel");
  const canStudents = admin || user.work_groups.some((g) => g.code === "academic");

  const items: NonNullable<TabsProps["items"]> = [];
  if (canPersonnel) items.push({ key: "personnel", label: "บุคลากร", icon: <TeamOutlined />, children: <PersonnelListPanel /> });
  if (canStudents) items.push({ key: "students", label: "นักเรียน/ผู้ปกครอง", icon: <UserOutlined />, children: <StudentsListPanel /> });

  return (
    <div className="flex flex-col gap-5">
      <PageHeader icon={<IdcardOutlined />} title="ข้อมูลบุคคล" subtitle="บุคลากร นักเรียน และผู้ปกครอง" />
      {items.length === 1 ? items[0]!.children : <Tabs items={items} />}
    </div>
  );
}
