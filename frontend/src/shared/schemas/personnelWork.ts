import { z } from "zod";

// schema ของผลงานครู (รายเทอม) + ไฟล์แนบ — ต้องตรงกับ backend
// (service.PersonnelWorkDTO / service.WorkFileDTO)

// ประเภทไฟล์แนบ — ตรงกับ backend domain (image|document|certificate)
export const workFileTypeSchema = z.enum(["image", "document", "certificate"]);
export type WorkFileType = z.infer<typeof workFileTypeSchema>;

export const workFileTypeLabel: Record<WorkFileType, string> = {
  image: "รูปภาพ",
  document: "เอกสาร",
  certificate: "เกียรติบัตร",
};

// --- ผลงาน ---
export const personnelWorkSchema = z.object({
  id: z.string(),
  title: z.string(),
  description: z.string(),
  work_date: z.string(),
  file_count: z.number(),
  created_at: z.string(),
});
export type PersonnelWorkRecord = z.infer<typeof personnelWorkSchema>;
export const personnelWorkListSchema = z.array(personnelWorkSchema);

export type WorkBody = {
  title: string;
  description: string;
  work_date: string;
};

// --- ไฟล์แนบ (url = signed URL หมดอายุ ตาม PDPA) ---
export const workFileSchema = z.object({
  id: z.string(),
  file_type: z.string(),
  original_name: z.string(),
  content_type: z.string(),
  size_bytes: z.number(),
  url: z.string(),
  created_at: z.string(),
});
export type WorkFileRecord = z.infer<typeof workFileSchema>;
export const workFileListSchema = z.array(workFileSchema);
