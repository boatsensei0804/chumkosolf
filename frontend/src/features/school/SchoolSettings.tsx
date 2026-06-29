"use client";

import { BankOutlined, CalendarOutlined, TeamOutlined } from "@ant-design/icons";
import { App, Button, Input, InputNumber, Spin } from "antd";
import Link from "next/link";
import { useState, type ReactNode } from "react";

import { AddressCascade } from "@/shared/ui/AddressCascade";
import { SectionCard } from "@/shared/ui/SectionCard";
import type { AddressData } from "@/shared/schemas/personnel";
import type { School, UpdateSchoolBody } from "@/shared/schemas/school";

import { useSchool, useUpdateSchool } from "./hooks";

// ===== ฟอร์มแก้ไขข้อมูลโรงเรียน (presentational) =====
export function SchoolInfoForm(props: {
  school: School;
  submitting: boolean;
  onSave: (body: UpdateSchoolBody) => void;
}): ReactNode {
  const { school, submitting, onSave } = props;
  const [name, setName] = useState(school.name);
  const [address, setAddress] = useState<AddressData>(school.address);
  const [phone, setPhone] = useState(school.phone);
  const [email, setEmail] = useState(school.email);
  const [website, setWebsite] = useState(school.website);
  const [directorName, setDirectorName] = useState(school.director_name);
  const [lateAfter, setLateAfter] = useState(school.attendance_late_after || "08:00");
  const [latePenalty, setLatePenalty] = useState<number>(school.attendance_late_penalty);
  const [error, setError] = useState("");

  const patchAddress = (patch: Partial<AddressData>): void => setAddress((prev) => ({ ...prev, ...patch }));

  const submit = (): void => {
    if (name.trim() === "") {
      setError("กรุณากรอกชื่อโรงเรียน");
      return;
    }
    if (!/^([01]\d|2[0-3]):[0-5]\d$/.test(lateAfter.trim())) {
      setError("เวลามาสายต้องอยู่ในรูปแบบ HH:MM (เช่น 08:00)");
      return;
    }
    if (email.trim() !== "" && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setError("รูปแบบอีเมลไม่ถูกต้อง");
      return;
    }
    if (latePenalty < 0 || latePenalty > 100 || Number.isNaN(latePenalty)) {
      setError("คะแนนหักเมื่อมาสายต้องอยู่ระหว่าง 0–100");
      return;
    }
    setError("");
    onSave({
      name: name.trim(),
      address,
      phone: phone.trim(),
      email: email.trim(),
      website: website.trim(),
      director_name: directorName.trim(),
      attendance_late_after: lateAfter.trim(),
      attendance_late_penalty: latePenalty,
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">
            ชื่อโรงเรียน<span className="text-red-500"> *</span>
          </label>
          <Input value={name} onChange={(e) => setName(e.target.value)} status={error ? "error" : ""} aria-label="ชื่อโรงเรียน" />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">รหัสโรงเรียน</label>
          <Input value={school.code} disabled aria-label="รหัสโรงเรียน" />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">ชื่อผู้อำนวยการ</label>
          <Input value={directorName} onChange={(e) => setDirectorName(e.target.value)} maxLength={150} aria-label="ชื่อผู้อำนวยการ" />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">เวลาเข้าเรียน (มาสายหลังเวลานี้)</label>
          <Input value={lateAfter} onChange={(e) => setLateAfter(e.target.value)} placeholder="08:00" maxLength={5} aria-label="เวลามาสาย" />
          <p className="mt-1 text-xs text-slate-400">รูปแบบ HH:MM — เช็คชื่อเข้าเรียนหลังเวลานี้จะถูกบันทึกเป็น “สาย”</p>
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">คะแนนหักเมื่อมาสาย</label>
          <InputNumber
            value={latePenalty}
            onChange={(v) => setLatePenalty(typeof v === "number" ? v : 0)}
            min={0}
            max={100}
            className="w-full"
            aria-label="คะแนนหักเมื่อมาสาย"
          />
          <p className="mt-1 text-xs text-slate-400">หักคะแนนความประพฤติอัตโนมัติเมื่อสแกนหน้าแล้วมาสาย (0 = ไม่หัก)</p>
        </div>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-slate-700">ที่อยู่</div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">บ้านเลขที่</label>
            <Input value={address.house_no} onChange={(e) => patchAddress({ house_no: e.target.value })} aria-label="บ้านเลขที่" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">หมู่</label>
            <Input value={address.moo} onChange={(e) => patchAddress({ moo: e.target.value })} aria-label="หมู่" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">ถนน</label>
            <Input value={address.road} onChange={(e) => patchAddress({ road: e.target.value })} aria-label="ถนน" />
          </div>
        </div>
        <div className="mt-4">
          <AddressCascade
            value={{
              province: address.province,
              district: address.district,
              subdistrict: address.subdistrict,
              postalCode: address.postal_code,
            }}
            onChange={(g) =>
              patchAddress({ province: g.province, district: g.district, subdistrict: g.subdistrict, postal_code: g.postalCode })
            }
          />
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">เบอร์โทร</label>
          <Input value={phone} onChange={(e) => setPhone(e.target.value)} maxLength={20} />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">อีเมล</label>
          <Input value={email} onChange={(e) => setEmail(e.target.value)} maxLength={150} placeholder="school@example.ac.th" aria-label="อีเมล" />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">เว็บไซต์</label>
          <Input value={website} onChange={(e) => setWebsite(e.target.value)} maxLength={255} placeholder="https://…" aria-label="เว็บไซต์" />
        </div>
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}
      <div className="flex justify-end">
        <Button type="primary" loading={submitting} onClick={submit}>
          บันทึกข้อมูลโรงเรียน
        </Button>
      </div>
    </div>
  );
}

// ลิงก์ไปยังส่วนตั้งค่าอื่น
function LinkCard(props: { icon: ReactNode; title: string; description: string; href?: string }): ReactNode {
  const inner = (
    <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white p-4 transition-colors hover:border-sky-200 hover:bg-sky-50/40">
      <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-sky-50 text-sky-600">{props.icon}</span>
      <div className="min-w-0">
        <div className="font-medium text-slate-800">{props.title}</div>
        <div className="text-xs text-slate-400">{props.description}</div>
      </div>
    </div>
  );
  return props.href ? <Link href={props.href}>{inner}</Link> : <div className="opacity-60">{inner}</div>;
}

export function SchoolSettings(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading, isError, error } = useSchool();
  const updateMutation = useUpdateSchool();

  return (
    <div className="flex flex-col gap-5">
      <SectionCard icon={<BankOutlined />} title="ข้อมูลโรงเรียน" description="ชื่อ ที่อยู่ และเบอร์ติดต่อของโรงเรียน">
        {isLoading ? (
          <Spin />
        ) : isError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
            โหลดข้อมูลไม่สำเร็จ: {error?.message}
          </div>
        ) : data ? (
          <SchoolInfoForm
            school={data}
            submitting={updateMutation.isPending}
            onSave={(body) =>
              updateMutation.mutate(body, {
                onSuccess: () => message.success("บันทึกข้อมูลโรงเรียนแล้ว"),
                onError: (err) => message.error(err.message),
              })
            }
          />
        ) : null}
      </SectionCard>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <LinkCard
          icon={<CalendarOutlined />}
          title="ปีการศึกษา / ภาคเรียน"
          description="กำหนดปี/เทอม และเทอมปัจจุบัน (อยู่ในเมนูตารางสอน)"
          href="/timetable"
        />
        <LinkCard icon={<TeamOutlined />} title="กลุ่มงาน / ผู้ใช้" description="จัดการสมาชิกกลุ่มงานและบัญชีผู้ใช้ (เร็ว ๆ นี้)" />
      </div>
    </div>
  );
}
