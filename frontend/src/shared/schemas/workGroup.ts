import { z } from "zod";

// schema กลุ่มงาน + การมอบหมาย — ตรงกับ backend (domain.WorkGroup / WorkGroupMembership)

export const workGroupSchema = z.object({
  id: z.string(),
  code: z.string(),
  name: z.string(),
});
export type WorkGroup = z.infer<typeof workGroupSchema>;
export const workGroupListSchema = z.array(workGroupSchema);

export const workGroupMembershipSchema = z.object({
  work_group_id: z.string(),
  code: z.string(),
  name: z.string(),
  is_group_admin: z.boolean(),
});
export type WorkGroupMembership = z.infer<typeof workGroupMembershipSchema>;
export const workGroupMembershipListSchema = z.array(workGroupMembershipSchema);

export type AssignWorkGroupBody = {
  work_group_id: string;
  is_group_admin: boolean;
};
