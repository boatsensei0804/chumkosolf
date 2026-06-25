import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { messageSchema } from "@/shared/schemas/personnel";
import {
  workGroupListSchema,
  workGroupMembershipListSchema,
  type AssignWorkGroupBody,
  type WorkGroup,
  type WorkGroupMembership,
} from "@/shared/schemas/workGroup";

function requireToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  }
  return token;
}

// list กลุ่มงานทั้งหมดของโรงเรียน
export async function listWorkGroups(): Promise<WorkGroup[]> {
  return apiRequest("/work-groups", workGroupListSchema, { token: requireToken() });
}

// กลุ่มงานที่บุคลากรสังกัด
export async function listPersonnelWorkGroups(personnelId: string): Promise<WorkGroupMembership[]> {
  return apiRequest(`/personnel/${personnelId}/work-groups`, workGroupMembershipListSchema, {
    token: requireToken(),
  });
}

export async function assignWorkGroup(personnelId: string, body: AssignWorkGroupBody): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/work-groups`, messageSchema, {
    method: "POST",
    body,
    token: requireToken(),
  });
}

export async function unassignWorkGroup(personnelId: string, workGroupId: string): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/work-groups/${workGroupId}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}
