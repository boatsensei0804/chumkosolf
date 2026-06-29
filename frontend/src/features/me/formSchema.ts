import { z } from "zod";

import { addressFormSchema, type AddressFormValues } from "@/features/personnel/formSchema";
import type { AddressData } from "@/shared/schemas/personnel";
import type { MyProfile, UpdateMyProfileBody } from "@/shared/schemas/me";

// form schema ของ "ข้อมูลของฉัน" — แก้ได้เฉพาะข้อมูลส่วนตัว/ที่อยู่
// (เลขบัตร/role/username เป็นงานกลุ่มบุคคล ไม่ให้แก้เอง)

const optionalEmail = z
  .string()
  .refine((v) => v === "" || z.string().email().safeParse(v).success, {
    message: "อีเมลไม่ถูกต้อง",
  });

export const myProfileFormSchema = z.object({
  prefix: z.string(),
  firstName: z.string().min(1, "กรุณากรอกชื่อ"),
  lastName: z.string().min(1, "กรุณากรอกนามสกุล"),
  birthDate: z.string(),
  phone: z.string(),
  email: optionalEmail,
  address: addressFormSchema,
});
export type MyProfileFormValues = z.infer<typeof myProfileFormSchema>;

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

export function toUpdateMyProfileBody(v: MyProfileFormValues): UpdateMyProfileBody {
  return {
    prefix: v.prefix,
    first_name: v.firstName,
    last_name: v.lastName,
    birth_date: v.birthDate,
    phone: v.phone,
    email: v.email,
    address: mapAddress(v.address),
  };
}

// แปลงข้อมูลจาก backend → ค่าเริ่มต้นของฟอร์ม
export function toMyProfileFormValues(d: MyProfile): MyProfileFormValues {
  return {
    prefix: d.prefix,
    firstName: d.first_name,
    lastName: d.last_name,
    birthDate: d.birth_date,
    phone: d.phone,
    email: d.email,
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
