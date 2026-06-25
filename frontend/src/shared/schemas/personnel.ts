import { z } from "zod";

// schema ของ personnel — ต้องตรงกับ backend (service.PersonnelListItem / PersonnelDetail)
// เลขบัตรประชาชนมาเป็น masked เท่านั้น (PDPA) ไม่มีเลขเต็มใน response

// ตำแหน่งบุคลากร: ครู/ผู้บริหาร (subset ของ user role)
export const personnelRoleSchema = z.enum(["teacher", "executive"]);
export type PersonnelRole = z.infer<typeof personnelRoleSchema>;

export const personnelRoleLabel: Record<PersonnelRole, string> = {
  teacher: "ครู",
  executive: "ผู้บริหาร",
};

export const addressSchema = z.object({
  house_no: z.string(),
  moo: z.string(),
  road: z.string(),
  subdistrict: z.string(),
  district: z.string(),
  province: z.string(),
  postal_code: z.string(),
});
export type AddressData = z.infer<typeof addressSchema>;

// รายการย่อสำหรับตาราง
export const personnelListItemSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  username: z.string(),
  role: z.string(),
  is_active: z.boolean(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  phone: z.string(),
  created_at: z.string(),
});
export type PersonnelListItem = z.infer<typeof personnelListItemSchema>;

export const personnelListSchema = z.array(personnelListItemSchema);

// รายละเอียดเต็ม (เลขบัตรยัง masked)
export const personnelDetailSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  username: z.string(),
  role: z.string(),
  is_active: z.boolean(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  civil_servant_id_masked: z.string(),
  birth_date: z.string(),
  phone: z.string(),
  email: z.string(),
  address: addressSchema,
  photo_path: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type PersonnelDetail = z.infer<typeof personnelDetailSchema>;

// response ของการสร้าง (คืน id)
export const createdIdSchema = z.object({ id: z.string() });

// response ข้อความ (update/delete)
export const messageSchema = z.object({ message: z.string() });

// payload ที่ส่งไป backend (snake_case ตาม JSON contract)
export type PersonnelAddressBody = AddressData;

export type CreatePersonnelBody = {
  username: string;
  password: string;
  role: PersonnelRole;
  national_id: string;
  civil_servant_id: string;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  email: string;
  address: PersonnelAddressBody;
};

export type UpdatePersonnelBody = {
  national_id: string;
  civil_servant_id: string;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  email: string;
  address: PersonnelAddressBody;
};
