"use client";

import { ScheduleOutlined } from "@ant-design/icons";
import { Tabs } from "antd";
import type { ReactNode } from "react";

import { AcademicManager } from "@/features/academic/AcademicManager";
import { useAuth } from "@/features/auth/AuthContext";
import { canManageTimetable, isSchoolAdmin } from "@/features/navigation/menu";
import { AssignmentsTab } from "@/features/timetable/AssignmentsTab";
import { ConfigTab } from "@/features/timetable/ConfigTab";
import { FreeTeachersTab } from "@/features/timetable/FreeTeachersTab";
import { GridTab } from "@/features/timetable/GridTab";
import { SubjectsTab } from "@/features/timetable/SubjectsTab";
import { PageHeader } from "@/shared/ui/PageHeader";

// แท็บจัดการปี/เทอม — เฉพาะผู้ดูแลระบบของโรงเรียน
function AcademicTermsTab(): ReactNode {
  const { user } = useAuth();
  if (!user || !isSchoolAdmin(user)) {
    return (
      <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700">
        การตั้งค่าปีการศึกษา/เทอม สำหรับผู้ดูแลระบบของโรงเรียนเท่านั้น
      </div>
    );
  }
  return <AcademicManager />;
}

export default function TimetablePage(): ReactNode {
  const { user } = useAuth();
  const canEdit = !!user && canManageTimetable(user);

  // ครูทั่วไป (ไม่ใช่กลุ่มวิชาการ) → ดูตารางห้องได้อย่างเดียว
  const items = canEdit
    ? [
        { key: "grid", label: "ตารางสอน", children: <GridTab /> },
        { key: "free", label: "ครูว่างวันนี้", children: <FreeTeachersTab /> },
        { key: "assignments", label: "มอบหมายสอน", children: <AssignmentsTab /> },
        { key: "subjects", label: "รายวิชา", children: <SubjectsTab /> },
        { key: "config", label: "ตั้งค่าคาบ", children: <ConfigTab /> },
        { key: "terms", label: "ปีการศึกษา/เทอม", children: <AcademicTermsTab /> },
      ]
    : [
        { key: "grid", label: "ตารางสอน (ดูอย่างเดียว)", children: <GridTab readOnly /> },
        { key: "free", label: "ครูว่างวันนี้", children: <FreeTeachersTab /> },
      ];

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<ScheduleOutlined />}
        title="ตารางสอน"
        subtitle={
          canEdit
            ? "กลุ่มงานวิชาการ · จัดรายวิชา/มอบหมายสอน/ตารางสอน"
            : "ดูตารางสอนของห้องเรียน (จัดการได้เฉพาะกลุ่มวิชาการ)"
        }
      />
      <div className="rounded-xl border border-slate-200 bg-white px-4 pb-4">
        <Tabs items={items} />
      </div>
    </div>
  );
}
