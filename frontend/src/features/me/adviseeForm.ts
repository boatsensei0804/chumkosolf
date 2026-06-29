import { z } from "zod";

import { addressFormSchema, type AddressFormValues } from "@/features/personnel/formSchema";
import type { AddressData } from "@/shared/schemas/personnel";
import type { UpdateMyAdviseeBody } from "@/shared/schemas/me";
import type { StudentDetail } from "@/shared/schemas/student";

// form schema ของการแก้ข้อมูลนักเรียนที่ปรึกษา (ครูที่ปรึกษา)
// แก้ได้เฉพาะข้อมูลส่วนตัว/ที่อยู่ — รหัสนักเรียน/สถานะ/เลขบัตร เป็นงานทะเบียน

export const adviseeFormSchema = z.object({
  prefix: z.string(),
  firstName: z.string().min(1, "กรุณากรอกชื่อ"),
  lastName: z.string().min(1, "กรุณากรอกนามสกุล"),
  birthDate: z.string(),
  phone: z.string(),
  address: addressFormSchema,
});
export type AdviseeFormValues = z.infer<typeof adviseeFormSchema>;

function mapAddress(a: AddressFormValues): AddressData {
  return {
    house_no: a.houseNo,
    moo: a.moo,
    road: a.road,
    subdistrict: a.subdistrict,
    district: a.district,
    province: a.province,
    postal_code: a.postalCode,
  };
}

export function toUpdateMyAdviseeBody(v: AdviseeFormValues): UpdateMyAdviseeBody {
  return {
    prefix: v.prefix,
    first_name: v.firstName,
    last_name: v.lastName,
    birth_date: v.birthDate,
    phone: v.phone,
    address: mapAddress(v.address),
  };
}

export function toAdviseeFormValues(d: StudentDetail): AdviseeFormValues {
  return {
    prefix: d.prefix,
    firstName: d.first_name,
    lastName: d.last_name,
    birthDate: d.birth_date,
    phone: d.phone,
    address: {
      houseNo: d.address.house_no,
      moo: d.address.moo,
      road: d.address.road,
      subdistrict: d.address.subdistrict,
      district: d.address.district,
      province: d.address.province,
      postalCode: d.address.postal_code,
    },
  };
}
