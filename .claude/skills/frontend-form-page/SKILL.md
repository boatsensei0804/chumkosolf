---
name: frontend-form-page
description: สร้างหน้าและฟอร์มใน frontend Next.js ของระบบโรงเรียนด้วย Ant Design + Tailwind + react-hook-form + zod แบบ TypeScript non-any ใช้ skill นี้เสมอเมื่อต้องสร้างหน้า dashboard, ฟอร์มกรอก/แก้ไขข้อมูล (นักเรียน ครู เช็คชื่อ คะแนน), ตารางข้อมูล หรือ component UI ใด ๆ แม้ผู้ใช้จะพูดแค่ "ทำหน้า X" หรือ "เพิ่มฟอร์ม Y" เพราะหน้า/ฟอร์มที่ไม่ทำตามแพทเทิร์น (ใช้ any, ไม่ validate ด้วย zod, ไม่ responsive, business logic ปนใน component) จะทำให้ระบบไม่สอดคล้องและแก้ยาก
---

# Frontend Form & Page

ทุกหน้าและฟอร์มทำตามแพทเทิร์นเดียวกัน: antd เป็น component, Tailwind จัด layout/responsive, react-hook-form + zod จัดการ form และ validation, TypeScript ห้าม any (ใช้คู่กับ frontend/CLAUDE.md)

## แพทเทิร์นฟอร์ม (บังคับทุกฟอร์ม)

ลำดับเสมอ: zod schema → infer type → useForm + zodResolver → antd ผ่าน Controller → error ไทยใต้ field

```tsx
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Input, Select, Button } from "antd";

const studentSchema = z.object({
  studentCode: z.string().min(1, "กรุณากรอกรหัสนักเรียน"),
  fullName: z.string().min(1, "กรุณากรอกชื่อ-นามสกุล"),
  classId: z.string().uuid("กรุณาเลือกห้อง"),
});
type StudentFormValues = z.infer<typeof studentSchema>;

export function StudentForm({ onSubmit }: { onSubmit: (v: StudentFormValues) => void }) {
  const { control, handleSubmit, formState: { errors, isSubmitting } } =
    useForm<StudentFormValues>({ resolver: zodResolver(studentSchema) });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
      <div>
        <label className="mb-1 block text-sm">รหัสนักเรียน</label>
        <Controller
          name="studentCode"
          control={control}
          render={({ field }) => <Input {...field} status={errors.studentCode ? "error" : ""} />}
        />
        {errors.studentCode && (
          <p className="mt-1 text-sm text-red-500">{errors.studentCode.message}</p>
        )}
      </div>
      {/* field อื่นทำแบบเดียวกัน */}
      <Button type="primary" htmlType="submit" loading={isSubmitting}>
        บันทึกข้อมูลนักเรียน
      </Button>
    </form>
  );
}
```

กฎ:
- antd ทุกตัวที่เป็น input ต้องห่อด้วย `Controller` (ไม่ใช่ register ตรง ๆ)
- ข้อความ error เป็นไทย มาจาก zod
- type มาจาก `z.infer` ไม่เขียน type มือซ้ำ
- ปุ่ม submit บอกชัดว่าทำอะไร + มี loading state
- enum ที่ใช้ร่วม (เช่นสถานะการลา) import จาก schema กลาง ไม่นิยามซ้ำ

## แพทเทิร์นหน้า Dashboard

- ใช้ layout กลาง: sidebar (เมนูตาม role/กลุ่มงาน) + topbar + content
- เมนูกรองตามสิทธิ์ผู้ใช้ — ไม่ render เมนูที่ไม่มีสิทธิ์
- การ์ดสรุปตัวเลขไว้บนสุด (เช่น จำนวนเช็คชื่อแล้ว/ขาด/ลา)
- ข้อมูลทุกหน้าผูกกับเทอมปัจจุบัน (semester จาก context) ส่งไปกับ request

```tsx
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
  <SummaryCard title="เช็คชื่อแล้ว" value={checked} />
  <SummaryCard title="ขาด" value={absent} />
  {/* ... */}
</div>
```

## พาเลตสี — ฟ้า-ขาว (บังคับ)

ใช้โทนฟ้า-ขาวทั้งระบบ กำหนดเป็น design token กลาง (Tailwind config + antd ConfigProvider) ห้าม hardcode hex กระจาย:
- `#2200FF` สีหลักเข้ม (ปุ่มหลัก/จุดเน้นสำคัญ)
- `#015FEB` สีหลัก (antd colorPrimary, link, active)
- `#0184FF` สีหลักสว่าง (hover, accent)
- `#00D4EB` สีเน้นรอง (highlight, badge, กราฟ)
- `#FFFFFF` พื้นหลัง/การ์ด

ขาวเป็นพื้นหลัก ฟ้าเป็น action/accent รักษา contrast ให้อ่านง่าย ตั้ง `ConfigProvider theme={{ token: { colorPrimary: '#015FEB' } }}`

## UX-first

ออกแบบจากการใช้งานจริงของครู/นักเรียนก่อน:
- งานที่ทำบ่อย (เช็คชื่อ, ดูคะแนน) เข้าถึงเร็ว กดน้อยครั้ง
- ลดภาระกรอก: default สมเหตุผล, เลือกแทนพิมพ์
- feedback ครบ: loading/success/error ทุก action
- ปุ่ม/ข้อความ active voice ภาษาไทย บอกชัดว่าทำอะไร
- error/empty state บอกทางออก ไม่ปล่อยว่าง

## Responsive (บังคับ)

- ใช้งานได้ตั้งแต่ ~375px ถึง desktop
- ใช้ Tailwind breakpoint สม่ำเสมอ (`sm md lg xl`)
- sidebar ยุบเป็น drawer บนจอเล็ก
- antd `Table` ใส่ `scroll={{ x: ... }}` เมื่อคอลัมน์เยอะ หรือสลับเป็น card list บนมือถือ
- ทดสอบทั้งจอเล็กและใหญ่ก่อนถือว่าเสร็จ

## Business logic แยกจาก component

- การเรียก API อยู่ใน custom hook / service layer ไม่เขียนใน component
- component โฟกัสที่ render + ผูก event
- ตัวอย่าง: `useStudents(semesterId)` คืน data + loading + error, component แค่เอาไปแสดง

```tsx
function StudentListPage() {
  const { semesterId } = useCurrentSemester();
  const { data, isLoading } = useStudents(semesterId); // logic อยู่ใน hook
  return <StudentTable data={data} loading={isLoading} />;
}
```

## TypeScript — non-any + สอดคล้องกับ backend (บังคับ)

- ห้าม `any` ทุกกรณี — ไม่รู้ type ใช้ `unknown` + narrow ด้วย type guard
- ทุก export function, props, state, ค่า return มี type ชัดเจน ไม่มี implicit any
- ห้าม `@ts-ignore`/`@ts-expect-error` เว้นแต่มีเหตุผล + comment
- type ข้อมูล API มาจาก zod schema (`z.infer`) ไม่เขียน type มือซ้ำ

**Type ต้องตรงกับ backend เสมอ:**
- ชื่อ field, โครงสร้าง, enum ต้องตรงกับที่ backend ส่ง/รับจริง
- enum ที่ใช้ร่วมกับ backend (เช่นสถานะการลา `present|absent|late|sick_leave|personal_leave`, relationship `father|mother|other`, ตำแหน่ง `director|deputy_director`) ต้องตรงเป๊ะ เก็บใน shared schema ที่เดียว
- response wrapper ตรงกับ backend: `{ success, data, error, meta }`
- backend เปลี่ยน contract → อัปเดต zod schema frontend ทันที ห้ามปล่อยหลุด type
- `npm run type-check` ต้องผ่าน 0 error ก่อนเสร็จ

## การทดสอบ (บังคับหลังเสร็จ)

- `npm run type-check` — ไม่มี any/type error
- `npm run lint` — ผ่าน
- `npm run test` — ผ่าน
- feature ใหม่มี test: render ได้, validation ทำงาน (กรอกผิดขึ้น error ไทย), submit เรียก API ถูก
- ทดสอบ responsive จอเล็ก/ใหญ่

## Checklist

- [ ] ฟอร์มใช้ zod schema → infer type → rhf + Controller
- [ ] error ภาษาไทยใต้ field
- [ ] ไม่มี any, type-check ผ่าน
- [ ] responsive ทดสอบจอเล็ก/ใหญ่แล้ว
- [ ] business logic/API อยู่ใน hook/service ไม่ปนใน component
- [ ] เมนู/หน้ากรองตามสิทธิ์ผู้ใช้
- [ ] ผูก semester ปัจจุบัน
- [ ] มี test ครบ + ผ่านทั้งหมด
