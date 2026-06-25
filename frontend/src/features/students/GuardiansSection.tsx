"use client";

import { DeleteOutlined, TeamOutlined } from "@ant-design/icons";
import { App, Button, Checkbox, Empty, Input, Popconfirm, Select, Spin, Tag } from "antd";
import { useState, type ReactNode } from "react";

import type { GuardianRelationship } from "@/shared/schemas/enums";
import { relationshipLabel, type LinkGuardianBody } from "@/shared/schemas/guardian";
import { SectionCard } from "@/shared/ui/SectionCard";

import { useLinkGuardian, useStudentGuardians, useUnlinkGuardian } from "./hooks";

const EMPTY_ADDRESS = {
  house_no: "",
  moo: "",
  road: "",
  subdistrict: "",
  district: "",
  province: "",
  postal_code: "",
};

// ===== ฟอร์มเพิ่มผู้ปกครอง (สร้าง inline ในหน้านักเรียน) — presentational =====
export function GuardianInlineForm(props: {
  onAdd: (body: LinkGuardianBody) => void;
  submitting: boolean;
}): ReactNode {
  const { onAdd, submitting } = props;
  const [nationalId, setNationalId] = useState("");
  const [prefix, setPrefix] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [phone, setPhone] = useState("");
  const [relationship, setRelationship] = useState<GuardianRelationship>("father");
  const [isPrimary, setIsPrimary] = useState(false);
  const [errors, setErrors] = useState<{ nationalId?: string; firstName?: string; lastName?: string }>({});

  const submit = (): void => {
    const next: typeof errors = {};
    if (!/^[0-9]{13}$/.test(nationalId)) next.nationalId = "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก";
    if (firstName.trim() === "") next.firstName = "กรุณากรอกชื่อ";
    if (lastName.trim() === "") next.lastName = "กรุณากรอกนามสกุล";
    setErrors(next);
    if (Object.keys(next).length > 0) return;

    onAdd({
      national_id: nationalId,
      prefix: prefix.trim(),
      first_name: firstName.trim(),
      last_name: lastName.trim(),
      birth_date: "",
      phone: phone.trim(),
      address: { ...EMPTY_ADDRESS },
      relationship,
      is_primary: isPrimary,
    });
    setNationalId("");
    setPrefix("");
    setFirstName("");
    setLastName("");
    setPhone("");
    setIsPrimary(false);
    setErrors({});
  };

  return (
    <div className="flex flex-col gap-3">
      <p className="text-xs font-medium text-slate-500">เพิ่มผู้ปกครอง</p>
      <div className="grid grid-cols-1 gap-x-3 gap-y-2.5 sm:grid-cols-2">
        <div>
          <Input value={prefix} onChange={(e) => setPrefix(e.target.value)} placeholder="คำนำหน้า" aria-label="คำนำหน้า" />
        </div>
        <div>
          <Input
            value={nationalId}
            onChange={(e) => setNationalId(e.target.value)}
            placeholder="เลขบัตรประชาชน 13 หลัก"
            aria-label="เลขบัตรประชาชน"
            maxLength={13}
            status={errors.nationalId ? "error" : ""}
          />
          {errors.nationalId && <p className="mt-1 text-xs text-red-500">{errors.nationalId}</p>}
        </div>
        <div>
          <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} placeholder="ชื่อ" aria-label="ชื่อ" status={errors.firstName ? "error" : ""} />
          {errors.firstName && <p className="mt-1 text-xs text-red-500">{errors.firstName}</p>}
        </div>
        <div>
          <Input value={lastName} onChange={(e) => setLastName(e.target.value)} placeholder="นามสกุล" aria-label="นามสกุล" status={errors.lastName ? "error" : ""} />
          {errors.lastName && <p className="mt-1 text-xs text-red-500">{errors.lastName}</p>}
        </div>
        <div>
          <Input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="เบอร์โทร" aria-label="เบอร์โทร" maxLength={20} />
        </div>
        <div>
          <Select<GuardianRelationship>
            value={relationship}
            onChange={setRelationship}
            className="w-full"
            aria-label="ความสัมพันธ์"
            options={[
              { value: "father", label: relationshipLabel.father },
              { value: "mother", label: relationshipLabel.mother },
              { value: "other", label: relationshipLabel.other },
            ]}
          />
        </div>
      </div>
      <div className="flex items-center justify-between">
        <Checkbox checked={isPrimary} onChange={(e) => setIsPrimary(e.target.checked)}>
          ผู้ปกครองหลัก
        </Checkbox>
        <Button type="primary" loading={submitting} onClick={submit}>
          เพิ่มผู้ปกครอง
        </Button>
      </div>
    </div>
  );
}

function relText(r: string): string {
  return r === "father" || r === "mother" || r === "other" ? relationshipLabel[r] : r;
}

// ===== section ผู้ปกครอง (container) =====
export function GuardiansSection({ studentId }: { studentId: string }): ReactNode {
  const { message } = App.useApp();
  const { data: links, isLoading } = useStudentGuardians(studentId);
  const linkMutation = useLinkGuardian(studentId);
  const unlinkMutation = useUnlinkGuardian(studentId);

  const handleAdd = (body: LinkGuardianBody): void => {
    linkMutation.mutate(body, {
      onSuccess: () => message.success("เพิ่มผู้ปกครองแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  return (
    <SectionCard icon={<TeamOutlined />} title="ผู้ปกครอง" description="เพิ่มผู้ปกครองและระบุผู้ปกครองหลัก" accent="emerald">
      {isLoading ? (
        <Spin />
      ) : (links?.length ?? 0) === 0 ? (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="ยังไม่มีผู้ปกครอง" />
      ) : (
        <ul className="divide-y divide-slate-100">
          {(links ?? []).map((l) => (
            <li key={l.id} className="flex items-center justify-between gap-3 py-3">
              <div className="flex flex-wrap items-center gap-2">
                <span className="font-medium text-slate-700">
                  {l.prefix}
                  {l.first_name} {l.last_name}
                </span>
                <Tag color="blue" bordered={false}>
                  {relText(l.relationship)}
                </Tag>
                {l.is_primary && (
                  <Tag color="gold" bordered={false}>
                    ผู้ปกครองหลัก
                  </Tag>
                )}
                <span className="num text-xs text-slate-400">{l.national_id_masked}</span>
              </div>
              <Popconfirm
                title="ถอดผู้ปกครองนี้?"
                okText="ถอด"
                cancelText="ยกเลิก"
                okButtonProps={{ danger: true }}
                onConfirm={() =>
                  unlinkMutation.mutate(l.id, {
                    onSuccess: () => message.success("ถอดผู้ปกครองแล้ว"),
                    onError: (err) => message.error(err.message),
                  })
                }
              >
                <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ถอด" />
              </Popconfirm>
            </li>
          ))}
        </ul>
      )}
      <div className="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
        <GuardianInlineForm onAdd={handleAdd} submitting={linkMutation.isPending} />
      </div>
    </SectionCard>
  );
}
