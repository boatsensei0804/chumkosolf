"use client";

import { Input, Select } from "antd";
import type { ReactNode } from "react";

import { useThaiAddress } from "@/shared/data/thaiAddress";

// ส่วนภูมิศาสตร์ของที่อยู่ (เลือกแบบ cascading + เติมรหัสไปรษณีย์อัตโนมัติ)
export type AddressGeo = {
  province: string;
  district: string; // อำเภอ/เขต
  subdistrict: string; // ตำบล/แขวง
  postalCode: string;
};

function Field(props: { label: string; children: ReactNode }): ReactNode {
  return (
    <div>
      <label className="mb-1.5 block text-sm font-medium text-slate-700">{props.label}</label>
      {props.children}
    </div>
  );
}

// AddressCascade — เลือกจังหวัด → อำเภอ → ตำบล แล้วเติมรหัสไปรษณีย์ให้อัตโนมัติ
// คืน 4 ฟิลด์ (วางในกริดที่อยู่ได้เลย); house_no/หมู่/ถนน ยังเป็น text แยกในฟอร์ม
export function AddressCascade(props: {
  value: AddressGeo;
  onChange: (next: AddressGeo) => void;
}): ReactNode {
  const { value, onChange } = props;
  const { data, isLoading } = useThaiAddress();

  const provinces = data ? Object.keys(data) : [];
  const amphoes = data && value.province ? Object.keys(data[value.province] ?? {}) : [];
  const tambons =
    data && value.province && value.district ? Object.keys(data[value.province]?.[value.district] ?? {}) : [];

  const opts = (arr: string[]): { value: string; label: string }[] =>
    arr.map((x) => ({ value: x, label: x }));

  const pickProvince = (p: string): void => onChange({ province: p, district: "", subdistrict: "", postalCode: "" });
  const pickDistrict = (d: string): void => onChange({ ...value, district: d, subdistrict: "", postalCode: "" });
  const pickSubdistrict = (s: string): void => {
    const zip = data?.[value.province]?.[value.district]?.[s] ?? "";
    onChange({ ...value, subdistrict: s, postalCode: zip });
  };

  return (
    <>
      <Field label="จังหวัด">
        <Select
          showSearch
          className="w-full"
          loading={isLoading}
          value={value.province || undefined}
          placeholder="เลือกจังหวัด"
          optionFilterProp="label"
          options={opts(provinces)}
          onChange={pickProvince}
        />
      </Field>
      <Field label="อำเภอ/เขต">
        <Select
          showSearch
          className="w-full"
          value={value.district || undefined}
          placeholder={value.province ? "เลือกอำเภอ/เขต" : "เลือกจังหวัดก่อน"}
          disabled={!value.province}
          optionFilterProp="label"
          options={opts(amphoes)}
          onChange={pickDistrict}
        />
      </Field>
      <Field label="ตำบล/แขวง">
        <Select
          showSearch
          className="w-full"
          value={value.subdistrict || undefined}
          placeholder={value.district ? "เลือกตำบล/แขวง" : "เลือกอำเภอก่อน"}
          disabled={!value.district}
          optionFilterProp="label"
          options={opts(tambons)}
          onChange={pickSubdistrict}
        />
      </Field>
      <Field label="รหัสไปรษณีย์">
        <Input className="num" value={value.postalCode} readOnly placeholder="เติมอัตโนมัติจากตำบล" />
      </Field>
    </>
  );
}
