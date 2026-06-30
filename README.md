# chumkosoft

ระบบบริหารจัดการโรงเรียน (School Management System) — ดู context และกฎทั้งหมดใน [CLAUDE.md](./CLAUDE.md)

## Stack

- **Backend:** Go + Fiber, PostgreSQL (pgx), Redis, golang-migrate, JWT
- **Frontend:** Next.js (App Router) + TypeScript, Ant Design, Tailwind, react-hook-form + zod
- **Infra:** Docker Compose (frontend / backend / postgres / redis)

## Quick start (dev)

```bash
cp .env.example .env          # แก้ค่า secret ตามต้องการ
docker compose up --build     # postgres → redis → migrate → backend → frontend
```

- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/health , readiness: http://localhost:8080/ready
- API base: http://localhost:8080/api/v1

migration จะรันอัตโนมัติผ่าน service `migrate` ก่อน backend ขึ้น

### สร้าง super admin ตั้งต้น

```bash
cd backend
SEED_ADMIN_PASSWORD=yourpassword \
DATABASE_URL=postgres://chumkosoft:chumkosoft_dev_pwd@localhost:5432/chumkosoft?sslmode=disable \
make seed-admin
```

## โครงสร้าง

```
chumkosoft/
├── docker-compose.yml          # dev (hot reload)
├── docker-compose.prod.yml     # prod (image ที่ build แล้ว)
├── backend/
│   ├── cmd/{api,seed}/         # entrypoints
│   ├── internal/               # config, database, server, middleware, httputil, tenant, crypto
│   └── migrations/             # golang-migrate (Phase 1)
└── frontend/
    └── src/{app,lib,shared}/   # App Router, theme, shared zod schemas
```

## คำสั่งทดสอบ

```bash
# backend
cd backend && go test ./... && go vet ./...
# frontend
cd frontend && npm run type-check && npm run lint && npm run test
```
