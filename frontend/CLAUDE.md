# frontend/CLAUDE.md — chumko-platform Next.js Frontend

อ่านไฟล์นี้ก่อนแตะโค้ด frontend ทุกครั้ง ใช้คู่กับ `../CLAUDE.md` (context หลัก)

---

## 1. Stack

- **Next.js** (App Router) + **TypeScript**
- **Ant Design (antd)** — component หลัก
- **Tailwind CSS** — layout, spacing, responsive
- **react-hook-form** — ทุกฟอร์ม
- **zod** — validation schema + type inference
- **TanStack Query** — data fetching/cache (ระบุเมื่อตั้งโปรเจค)

---

## 2. TypeScript — ห้าม any เด็ดขาด (บังคับสูงสุด)

- **ห้ามใช้ `any`** ทุกกรณี ไม่รู้ type ใช้ `unknown` + narrow ด้วย type guard
- เปิด `strict: true` ห้ามปิด strict flags
- ห้าม `@ts-ignore` / `@ts-expect-error` เว้นแต่มีเหตุผลและ comment กำกับ
- ทุก export function มี return type ชัดเจน
- ทุก props, state, ตัวแปร, ค่า return ต้องมี type — ไม่มีอะไร implicit any
- `tsc --noEmit` (`npm run type-check`) ต้องผ่านก่อนเสร็จ

### Type ต้องสอดคล้องกับ backend
นี่สำคัญมาก — type ฝั่ง frontend ต้องตรงกับ API จริงของ backend เสมอ:
- Type ของ request/response มาจาก **zod schema** แล้ว infer (`z.infer<typeof schema>`) — ไม่เขียน type มือซ้ำ
- โครงสร้าง field, ชื่อ field, enum ต้องตรงกับที่ backend ส่ง/รับจริง (เช่น สถานะการลา `present|absent|late|sick_leave|personal_leave` ต้องตรงเป๊ะกับ backend)
- response wrapper ตรงกับ backend: `{ success, data, error, meta }`
- ถ้า backend เปลี่ยน contract ต้องอัปเดต zod schema ฝั่ง frontend ให้ตรงทันที ห้ามปล่อยให้หลุด type
- เก็บ shared schema/enum ที่ใช้ตรงกับ backend ไว้ที่เดียว (เช่น `src/shared/schemas/`) ไม่กระจาย

---

## 3. UX/UI — ยึด UX เป็นหลัก

### 3.1 หลักการ UX-first
ออกแบบโดยคิดจากการใช้งานจริงของครู/นักเรียนก่อนเสมอ ไม่ใช่คิดจากโครงสร้างข้อมูล:
- งานที่ทำบ่อย (เช็คชื่อ, ดูคะแนน) ต้องเข้าถึงเร็ว กดน้อยครั้ง
- ลดภาระการกรอก: ใช้ default ที่สมเหตุผล, จำค่าที่เลือกบ่อย, เลือกแทนพิมพ์เมื่อทำได้
- feedback ชัดทันที: loading, success, error state ครบทุก action
- ปุ่ม/ข้อความบอกชัดว่าทำอะไร (active voice ภาษาไทย): "บันทึกการเช็คชื่อ" ไม่ใช่ "Submit"
- ปุ่มกับ toast สอดคล้องกัน (กด "บันทึก" → "บันทึกแล้ว")
- หน้า error/empty ต้องบอกว่าเกิดอะไรและทำต่อยังไง ไม่ปล่อยว่าง

### 3.2 Responsive (บังคับ)
- ใช้งานได้ตั้งแต่มือถือ (~375px) ถึง desktop ทุกหน้า
- sidebar ยุบเป็น drawer บนจอเล็ก
- antd `Table` ใส่ `scroll={{ x }}` หรือสลับเป็น card list บนมือถือเมื่อคอลัมน์เยอะ
- ใช้ Tailwind breakpoint (sm/md/lg/xl) สม่ำเสมอ ทดสอบจอเล็ก+ใหญ่ก่อนเสร็จ

### 3.3 Dashboard เป็นรูปแบบหลัก
- layout: sidebar (เมนูตาม role/กลุ่มงาน) + topbar + content
- เมนูกรองตามสิทธิ์ — ไม่ render เมนูที่ผู้ใช้ไม่มีสิทธิ์
- การ์ดสรุปตัวเลขไว้บนสุด

---

## 4. พาเลตสี — ฟ้า-ขาวเป็นหลัก (บังคับ)

ทิศทางจาก skill `ui-ux-pro-max` (Government/Public Service · Data-Dense) — navy + professional blue + slate, dark sidebar:

| ชื่อ | HEX | = Tailwind | ใช้กับ |
|------|-----|-----------|--------|
| Primary | `#0369A1` | sky-700 | สีหลัก (link, active, ปุ่ม primary) |
| Primary deep | `#075985` | sky-800 | สีหลักเข้ม (hover, ปลาย gradient) |
| Primary bright | `#0EA5E9` | sky-500 | highlight |
| Navy | `#0F172A` | slate-900 | dark sidebar / พื้นผิวเข้ม |
| White | `#FFFFFF` | — | พื้นการ์ด/พื้นผิว |

> รสนิยม/ดีไซน์ทั้งหมดอ้างอิง `frontend/design-system/MASTER.md` (source of truth) ซึ่งยึด skill `ui-ux-pro-max` + `frontend-design` — รวม accent color-coding ต่อหมวด (`src/shared/ui/accent.ts`)

กฎการใช้สี:
- กำหนดสีเป็น **design token กลาง** (Tailwind config + antd ConfigProvider) ห้าม hardcode hex กระจายในแต่ละ component
- ตั้ง antd `theme.token.colorPrimary = '#0369A1'` ผ่าน `ConfigProvider`
- พื้นแอป `slate-50`, การ์ดขาว, **sidebar navy** เป็นจุดเด่น enterprise, ฟ้าเป็น accent/action
- contrast: บนพื้น navy ใช้ขาว/slate-300, บนขาวใช้ slate-800/900 — ตรวจ ≥ 4.5:1
- สี status (success `#059669` / warning amber / error `#DC2626`) ใช้กับ badge/แถว
- ⚠️ แก้ custom color ใน `tailwind.config.ts` ต้อง **restart dev server** (HMR ไม่ reload config)

---

## 5. Form — react-hook-form + zod ทุกฟอร์ม

ลำดับเสมอ: zod schema → infer type → useForm + zodResolver → antd ผ่าน Controller → error ไทยใต้ field

- antd input ทุกตัวห่อด้วย `Controller` (ไม่ register ตรง)
- error เป็นไทยจาก zod
- type จาก `z.infer` ไม่เขียนมือซ้ำ
- validation frontend ไม่ทดแทน backend — มีทั้งสองฝั่ง
- schema/enum ที่ใช้ร่วมกับ backend เก็บที่ shared folder
- ดู skill frontend-form-page

---

## 6. โครงสร้างและ Business Logic

- แยกชั้น: UI component / business logic (hooks, services) / API client
- business logic อยู่ใน custom hooks หรือ service layer ไม่ปนใน component การเรียก API ไม่เขียนตรงใน component
- ทุก feature ผูก school_id + semester_id จาก context/auth
- ห้ามสร้างหน้า/เมนู/ฟอร์มนอกขอบเขต business logic ในเอกสารหลัก ไม่แน่ใจให้ถาม

---

## 7. ข้อมูลส่วนบุคคล (PDPA)
- ไม่แสดงเลขบัตรประชาชนเต็มถ้าไม่จำเป็น (แสดงแบบ mask)
- ไฟล์แนบโหลดผ่าน signed URL ไม่ฝัง public URL
- ไม่ log ข้อมูลส่วนบุคคลลง console/error tracking

---

## 8. การทดสอบ (บังคับหลังเสร็จทุกครั้ง)
- `npm run type-check` — ไม่มี any/type error
- `npm run lint` — ผ่าน
- `npm run test` — ผ่าน
- feature ใหม่มี test: render ได้, validation ทำงาน (กรอกผิดขึ้น error ไทย), submit เรียก API ถูก
- ทดสอบ responsive จอเล็ก/ใหญ่
ถ้าข้อใดไม่ผ่าน ห้ามรายงานว่าเสร็จ

---

## 9. กฎการแก้โค้ด (ย้ำจาก context หลัก)
- ห้ามลบไฟล์โดยไม่จำเป็น
- ห้ามแก้โค้ดที่ไม่เกี่ยวข้อง (minimal diff)
- ห้าม reformat ทั้งไฟล์โดยไม่จำเป็น
- เห็นจุดควรปรับนอกขอบเขต → แจ้งผู้ใช้ ไม่แก้เอง
