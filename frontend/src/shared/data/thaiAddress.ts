"use client";

import { useQuery, type UseQueryResult } from "@tanstack/react-query";

// โครงสร้าง: จังหวัด → อำเภอ/เขต → ตำบล/แขวง → รหัสไปรษณีย์
export type ThaiAddressData = Record<string, Record<string, Record<string, string>>>;

async function loadThaiAddress(): Promise<ThaiAddressData> {
  const res = await fetch("/thai-address.json");
  if (!res.ok) throw new Error("โหลดข้อมูลที่อยู่ไม่สำเร็จ");
  return (await res.json()) as ThaiAddressData;
}

// useThaiAddress โหลดชุดข้อมูลที่อยู่ไทยครั้งเดียว (cache ถาวรในเซสชัน)
export function useThaiAddress(): UseQueryResult<ThaiAddressData, Error> {
  return useQuery({
    queryKey: ["thai-address"],
    queryFn: loadThaiAddress,
    staleTime: Infinity,
    gcTime: Infinity,
  });
}
