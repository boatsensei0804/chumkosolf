import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "antd";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudentPhotoCard } from "./StudentPhotoCard";
import type { StudentPhoto } from "@/shared/schemas/student";

let photos: StudentPhoto[] = [];
const uploadMutate = vi.fn();
const primaryMutate = vi.fn();
const deleteMutate = vi.fn();

vi.mock("./photoHooks", () => ({
  useStudentPhotos: () => ({ data: photos, isLoading: false }),
  useUploadStudentPhoto: () => ({ mutate: uploadMutate, isPending: false }),
  useSetStudentPhotoPrimary: () => ({ mutate: primaryMutate, isPending: false }),
  useDeleteStudentPhoto: () => ({ mutate: deleteMutate, isPending: false }),
}));

function renderCard(ui: ReactNode): ReturnType<typeof render> {
  return render(<App>{ui}</App>);
}

beforeEach(() => {
  uploadMutate.mockClear();
  primaryMutate.mockClear();
  deleteMutate.mockClear();
  photos = [];
});

describe("StudentPhotoCard (gallery)", () => {
  it("ยังไม่มีรูป → แสดง empty + ปุ่มอัปโหลด/ถ่ายรูป", () => {
    renderCard(<StudentPhotoCard studentId="s1" />);
    expect(screen.getByText("ยังไม่มีรูปนักเรียน")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /อัปโหลดรูป/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ถ่ายรูป/ })).toBeInTheDocument();
  });

  it("มีหลายรูป → รูปโปรไฟล์มี badge และไม่มีปุ่ม 'ตั้งเป็นรูปโปรไฟล์', รูปอื่นมีปุ่ม", () => {
    photos = [
      { id: "p1", url: "https://x/1.jpg", is_primary: true, created_at: "" },
      { id: "p2", url: "https://x/2.jpg", is_primary: false, created_at: "" },
    ];
    renderCard(<StudentPhotoCard studentId="s1" />);
    expect(screen.getByText("โปรไฟล์")).toBeInTheDocument();
    expect(screen.getAllByRole("img", { name: "รูปนักเรียน" })).toHaveLength(2);
    // รูปที่ไม่ใช่โปรไฟล์ → มีปุ่มตั้งโปรไฟล์ 1 ปุ่ม
    const setButtons = screen.getAllByRole("button", { name: "ตั้งเป็นรูปโปรไฟล์" });
    expect(setButtons).toHaveLength(1);
  });

  it("กดตั้งรูปโปรไฟล์ → เรียก mutate ด้วย id ของรูปนั้น", async () => {
    const user = userEvent.setup();
    photos = [
      { id: "p1", url: "https://x/1.jpg", is_primary: true, created_at: "" },
      { id: "p2", url: "https://x/2.jpg", is_primary: false, created_at: "" },
    ];
    renderCard(<StudentPhotoCard studentId="s1" />);
    await user.click(screen.getByRole("button", { name: "ตั้งเป็นรูปโปรไฟล์" }));
    expect(primaryMutate).toHaveBeenCalledWith("p2", expect.anything());
  });
});
