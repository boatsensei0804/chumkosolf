# chumko-platform — Design System (MASTER)

Source of truth สำหรับ UI ทั้งระบบ ทุกหน้า/ทุก component ต้องอิงไฟล์นี้

**ที่มาของรสนิยม/ดีไซน์:** ยึด skill **`ui-ux-pro-max`** + **`frontend-design`** เป็นหลัก (แทนกฎ design เดิมใน `frontend-form-page`) — ทิศทาง: Data-Dense Dashboard
**กฎวิศวกรรม/ความปลอดภัย** (TypeScript non-any, zod, **PDPA mask**, **tenant school_id scope**, business-logic-แยก, responsive) ยังคงตาม `frontend/CLAUDE.md` + skill เดิม — non-negotiable

**โทนรวม:** สะอาด เป็นทางการ น่าเชื่อถือ สำหรับครู/ผู้บริหารโรงเรียนไทย — ฟ้า-ขาว, อ่านง่าย, ไม่หวือหวา แต่ "ตั้งใจ" (ไม่ใช่ default antd ดิบ) ใช้ความเด่นที่จุดเดียว (hero/แบรนด์) ที่เหลือเงียบ

Stack คงที่ (ห้ามเปลี่ยน): **Next.js + Ant Design + Tailwind + react-hook-form + zod**, TypeScript non-any

---

## 1. สี (Color tokens)

ทิศทางจาก **ui-ux-pro-max → Data-Dense Dashboard** กำหนดเป็น token กลางที่ `tailwind.config.ts` (`brand.*`) และ `src/lib/theme.ts` (antd) — **ห้าม hardcode hex ใน component**

| Token | HEX | ใช้กับ |
|-------|-----|--------|
| `brand.DEFAULT` | `#2563EB` | primary: ปุ่ม, link, active, ไอคอนเน้น |
| `brand.deep` | `#1E40AF` | ปลาย gradient, จุดเน้นสุด |
| `brand.bright` | `#3B82F6` | secondary, hover, link hover |
| `brand.accent` | `#059669` | accent เขียว (CTA รอง, success เน้น) |
| `brand.cyan` | `#00D4EB` | decorative (วงตกแต่งใน hero) |
| white | `#FFFFFF` | พื้นการ์ด/พื้นผิว |

**Neutral (Tailwind slate):** พื้นแอป `slate-50 (#f8fafc)` · foreground `slate-900 (#0f172a)` · รอง `slate-500` · จาง `slate-400` · เส้น `slate-200`/`slate-100` (border รอง `#e4ecfc`)

**Semantic / status (data-dense):** success = `#059669` / antd green · error = `#DC2626` / `text-red-500` · warning = `amber-*` — ใช้สีสถานะกับ badge/แถว เพื่อให้ scan ง่าย

**กฎคอนทราสต์:** ตัวอักษรเนื้อหาใช้ `slate-800/900` (ไม่ใช่ฟ้า), `brand` เป็น action/accent เท่านั้น, ตัวอักษรบนพื้น gradient ฟ้าใช้ขาว (heading ใหญ่), ทุกคู่ ≥ 4.5:1

---

## 2. ตัวอักษร (Typography)

- ฟอนต์เนื้อหา/หัวข้อ: **Sarabun** (`--font-sans`, โหลดผ่าน `next/font`) — รองรับไทย เหมาะครู/ผู้บริหาร
- ฟอนต์ตัวเลข/โค้ด: **Fira Code** (`--font-mono`) — ใช้กับเลขบัตร/เบอร์/วันที่/KPI ผ่าน class `.num` (mono + tabular-nums) ให้ดู technical และกันการขยับ
- Scale: `text-xs 12 · sm 14 · base 16 · lg 18 · xl 20 · 2xl 24 · 3xl 30`
- น้ำหนัก: heading **700** (`font-bold`), หัวข้อย่อย/ปุ่ม/label **500-600**, เนื้อหา **400**
- หัวข้อ section ในฟอร์ม: เล็ก + `uppercase tracking-wide text-slate-500` + แถบ accent ฟ้าซ้าย (ดู §6)

---

## 3. ระยะห่าง / มุม / เงา / เส้น

- **Spacing:** ยึดจังหวะ 4/8px (`gap-2/3/4/5/6`, `p-4/5/6`) — section spacing `gap-5/6`
- **Radius:** ปุ่ม/อินพุต `rounded-lg`(antd 10) · การ์ด `rounded-xl`/`rounded-2xl` · pill `rounded-full` · โลโก้ `rounded-xl`
- **Shadow:** เบามาก — การ์ดใช้ `border border-slate-200` เป็นหลัก + เงาจาง (`shadow-sm`/hover `shadow-md`) ไม่ใช้เงาหนา
- **Border:** เส้นคั่น `border-slate-100`, ขอบการ์ด `border-slate-200`
- **Container:** เนื้อหาหน้า `max-w-6xl` (ฟอร์มเดี่ยว `max-w-4xl`) จัดกึ่งกลาง `mx-auto`

---

## 4. Layout หลัก

- **Dashboard shell** (`features/navigation/DashboardLayout`): sidebar ขาว 256px (เดสก์ท็อป) / Drawer (มือถือ) + topbar sticky blur + content `bg-slate-50 p-4 md:p-6`
- **Auth (login):** split-panel — แผงซ้าย gradient ฟ้า (แบรนด์+คุณค่า) / แผงขวา ฟอร์ม; มือถือซ่อนแผงซ้าย
- Responsive: ทำงานได้ ~375px → desktop, ใช้ breakpoint `sm md lg xl` สม่ำเสมอ, ไม่มี horizontal scroll

---

## 5. Signature elements (เอกลักษณ์)

1. **Brand mark:** โลโก้สี่เหลี่ยมมน `bg-gradient-to-br from-brand to-brand-deep` + ไอคอน `ReadOutlined` สีขาว + wordmark "ชุมโค / ระบบบริหารโรงเรียน"
2. **Hero gradient:** แถบ `from-brand to-brand-deep` มุมมน + วงกลมตกแต่ง (`bg-white/10`, `bg-brand-cyan/15`) — ใช้เป็นโมเมนต์เด่น **จุดเดียวต่อหน้า** (หน้าแรก, แผง login)
3. **Active nav bar:** เมนู active = พื้น `brand/10` + แถบ `bg-brand` ด้านซ้าย + ตัวหนา `text-brand-deep`

---

## 6. แพทเทิร์น Component

**Page header** (`shared/ui/PageHeader`): ลิงก์ย้อนกลับ (← ...) + ไอคอนกรอบ `bg-brand/10 text-brand` + title `text-xl font-bold` + subtitle `text-slate-500` + actions ฝั่งขวา

**Section card** (`shared/ui/SectionCard`): `rounded-2xl border-slate-200 bg-white`; header `border-b border-slate-100 px-5 py-4` (ไอคอน + title + description + extra); body `p-5`

**Section header ในฟอร์ม:** `border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500`

**Form field:** label `text-sm font-medium text-slate-700` (มองเห็นเสมอ ไม่ placeholder-only) → antd input ผ่าน `Controller` → error ไทยใต้ field `text-sm text-red-500`; required มี `*`; input `size="large"` ในฟอร์มหลัก/หน้า login

**Badge/Tag:** `Tag bordered={false}` — role: teacher=blue, executive=purple; สถานะ: ใช้งาน=success, ระงับ=default; "ปัจจุบัน"/"หัวหน้ากลุ่ม"=green/gold; pill "เร็ว ๆ นี้" = `bg-slate-100 text-slate-400 text-[10px] rounded-full`

**Row list** (sub-resource): `divide-y divide-slate-100`, แต่ละแถว `py-3 flex items-center justify-between`, ปุ่มลบเป็น icon button (`type="text" danger`) + `Popconfirm`

**Avatar:** วงกลม gradient `from-brand to-brand-deep` + ตัวอักษรแรกขาว (topbar) หรือ `bg-brand/10 text-brand` initial (ตาราง)

**Empty state:** ข้อความบอกทางออก + ปุ่มถ้าทำได้ (ไม่ปล่อยว่าง) เช่น "ยังไม่มีข้อมูล — กด …"

**Table:** ครอบในกรอบการ์ด (`rounded-xl border`), header `bg-slate-50`, `scroll={{x}}` เมื่อคอลัมน์เยอะ, action เป็น icon button ชิดขวา

**ปุ่ม:** primary = action หลัก 1 ปุ่ม/บริบท, ข้อความ active voice ("เพิ่มบุคลากร", "บันทึกการแก้ไข", "มอบหมาย"); ปุ่มกับ toast สอดคล้องกัน; มี loading state

---

## 7. UX / Accessibility / PDPA

- **Active voice ไทย** ทุกปุ่ม/ข้อความ; error บอกสาเหตุ+ทางแก้; empty/loading/error ครบทุก action
- feedback: loading (spinner/skeleton), success (toast), error (alert/toast ไทยจาก backend)
- a11y: label ผูก input, icon-only button มี `aria-label`, focus ring คงไว้, contrast ผ่าน, ไม่สื่อด้วยสีอย่างเดียว
- micro-interaction: transition 150–300ms (`transition-colors/all`), hover ยกการ์ด `-translate-y-0.5`
- **PDPA:** เลขบัตรประชาชนแสดง **mask** เสมอ (`1-2345-xxxxx-xx-1`), ตอนแก้ไขไม่ดึงเลขเต็ม, ไม่ log ข้อมูลส่วนบุคคล

---

## 8. Anti-patterns (อย่าทำ)

- ❌ ใช้ emoji เป็นไอคอน (ใช้ `@ant-design/icons`) · ❌ hardcode hex · ❌ เงาหนา/หลายระดับมั่ว
- ❌ ฟ้าทั้งหน้าจนล้า (ฟ้าเป็น accent, ขาว/slate เป็นพื้น) · ❌ ตัวอักษรเทาบนเทา
- ❌ placeholder แทน label · ❌ error รวมบนสุดอย่างเดียว (วางใต้ field) · ❌ หน้า empty ปล่อยว่าง
- ❌ เปลี่ยน stack เป็น shadcn/อื่น · ❌ business logic ใน component (อยู่ใน hook/service)

---

## อ้างอิงการใช้งานจริง

- tokens: `tailwind.config.ts`, `src/lib/theme.ts`, `src/app/globals.css`
- shell/nav: `src/features/navigation/DashboardLayout.tsx`
- shared ui: `src/shared/ui/`
- หน้าอ้างอิง: `src/app/login/page.tsx`, `src/app/(dashboard)/page.tsx`, `src/app/(dashboard)/personnel/*`
