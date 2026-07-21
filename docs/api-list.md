# API Documentation — Cloud Automation Portal

Base URL: `http://<host>:8080/api` — **mọi endpoint đều bắt đầu bằng `/api`**, kể cả
`/api/nodes` (không phải `/nodes`).

Tất cả response lỗi có dạng: `{ "error": "<mô tả>" }`

Trong các lệnh `curl` dưới đây, thay `localhost` bằng IP thật nếu gọi từ máy khác,
và thay `<ACCESS_TOKEN>` / `<REFRESH_TOKEN>` bằng giá trị thật lấy từ response `/login`.

---

## Auth

### POST /api/auth/login
Đăng nhập, nhận access token + refresh token.

**Auth required:** Không

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "TyyhJNXS31WBB0jz...",
  "expires_in": 1800,
  "user": {
    "id": "d1aa6423-4f05-4dd8-b1e1-c336426833b1",
    "username": "admin",
    "role": "admin",
    "created_at": "2026-07-20T10:13:37Z"
  }
}
```

**Response 401:** sai username/password (message chung, không tiết lộ username có tồn tại hay không)

---

### POST /api/auth/refresh
Cấp access token mới từ refresh token. Áp dụng **rotation**: refresh token cũ bị thu hồi ngay khi dùng, refresh token mới được cấp thay thế.

**Auth required:** Không (dùng refresh token thay vì access token)

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

**Response 200:**
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "<refresh_token_mới>",
  "expires_in": 1800
}
```

**Response 401:**
- Token không tồn tại / sai định dạng
- Token đã hết hạn (`REFRESH_TOKEN_TTL_DAYS` trong config)
- Token đã bị thu hồi (đã dùng để refresh trước đó, hoặc đã logout)

---

### POST /api/auth/logout
Thu hồi refresh token hiện tại — sau khi gọi, token này không dùng để `/refresh` được nữa.

**Auth required:** Có (Bearer access token)

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

**Response 200:**
```json
{ "message": "đăng xuất thành công" }
```

---

### GET /api/auth/me
Lấy thông tin user hiện tại (đọc trực tiếp từ claims trong JWT, không query DB).

**Auth required:** Có (Bearer access token)

```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

**Response 200:**
```json
{
  "id": "d1aa6423-4f05-4dd8-b1e1-c336426833b1",
  "username": "admin",
  "role": "admin"
}
```

---

### POST /api/auth/register
Tạo tài khoản mới. Chỉ admin được phép gọi.

**Auth required:** Có (Bearer access token, role = `admin`)

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>" \
  -d '{"username":"peano","password":"user12345","role":"user"}'
```
`role` chỉ nhận `admin` hoặc `user`.

**Response 201:**
```json
{
  "id": "...",
  "username": "peano",
  "role": "user",
  "created_at": "2026-07-20T..."
}
```

**Response 403:** người gọi không phải admin
**Response 409:** username đã tồn tại

---

## Node

### GET /api/nodes
Danh sách node Proxmox kèm % sử dụng CPU/RAM/Disk.

**Auth required:** Có (Bearer access token — mọi role)

```bash
curl -X GET http://localhost:8080/api/nodes \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

**Response 200:**
```json
[
  {
    "node": "pve-node1",
    "status": "online",
    "cpu_percent": 12.5,
    "maxcpu": 8,
    "mem_gib": 4.2,
    "maxmem_gib": 16,
    "mem_percent": 26.25,
    "disk_gib": 50.1,
    "maxdisk_gib": 200,
    "disk_percent": 25.05
  }
]
```

---

## VM

### GET /api/nodes/:node/vms
Danh sách VM trong 1 node.

**Auth required:** Có (mọi role)

```bash
curl -X GET http://localhost:8080/api/nodes/pve-node1/vms \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

**Response 200:**
```json
[
  {
    "vmid": 100,
    "name": "web-server-01",
    "status": "running",
    "cpu_percent": 5.3,
    "mem_gib": 2.1,
    "maxmem_gib": 4,
    "mem_percent": 52.5,
    "maxdisk_gib": 32
  }
]
```

---

### POST /api/nodes/:node/vms
Tạo VM mới.

**Auth required:** Có (mọi role)

```bash
curl -X POST http://localhost:8080/api/nodes/pve-node1/vms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{"vmid":101,"name":"test-vm","cores":2,"memory":2048,"ostype":"l26"}'
```

**Response 200:**
```json
{ "message": "Đang tạo VM", "task_id": "UPID:..." }
```

> **Lưu ý còn tồn đọng:** endpoint này hiện tạo VM mới hoàn toàn, **chưa phải** clone từ template như yêu cầu MVP gốc — cần rà lại thiết kế ở bước tiếp theo.

---

### DELETE /api/nodes/:node/vms/:vmid
Xóa VM.

**Auth required:** Có, **role = admin**

```bash
curl -X DELETE http://localhost:8080/api/nodes/pve-node1/vms/101 \
  -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>"
```

**Response 200:**
```json
{ "message": "Đang xóa VM", "task_id": "UPID:..." }
```

**Response 403:** người gọi không phải admin

---

### POST /api/nodes/:node/vms/:vmid/start, /stop, /shutdown, /reboot, /reset

Điều khiển vòng đời VM.

**Auth required:** Có (mọi role)

```bash
curl -X POST http://localhost:8080/api/nodes/pve-node1/vms/101/start \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

curl -X POST http://localhost:8080/api/nodes/pve-node1/vms/101/stop \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

curl -X POST http://localhost:8080/api/nodes/pve-node1/vms/101/shutdown \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

curl -X POST http://localhost:8080/api/nodes/pve-node1/vms/101/reboot \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

curl -X POST http://localhost:8080/api/nodes/pve-node1/vms/101/reset \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

**Response 200:**
```json
{ "message": "Lệnh đã được gửi", "task_id": "UPID:..." }
```

---

## Chuỗi lệnh test nhanh — copy chạy tuần tự (Bash / Ubuntu)

```bash
# 1. Login, lưu access_token và refresh_token vào biến
RESP=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

echo "$RESP"

ACCESS_TOKEN=$(echo "$RESP" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$RESP" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)

echo "Access: $ACCESS_TOKEN"
echo "Refresh: $REFRESH_TOKEN"

# 2. Gọi API được bảo vệ
curl -X GET http://localhost:8080/api/nodes \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Refresh
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"

# 4. Logout (dùng refresh_token MỚI nếu vừa refresh ở bước 3)
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

*(Cách này dùng `grep`/`cut` để tự trích JSON — không cần cài `jq`. Nếu VM có `jq`, gọn hơn với `echo "$RESP" | jq -r .access_token`)*

---

## Ghi chú triển khai

- Access token (JWT) hết hạn sau `ACCESS_TOKEN_TTL_MINUTES` (mặc định 30 phút)
- Refresh token hết hạn sau `REFRESH_TOKEN_TTL_DAYS` (mặc định 7 ngày)
- Mọi request tới route bảo vệ cần header: `Authorization: Bearer <access_token>`
- **Mọi endpoint đều có tiền tố `/api`** — lỗi `404 page not found` phổ biến nhất là do quên tiền tố này
- Role hiện có: `admin`, `user` — mở rộng RBAC chi tiết hơn để ở phase sau MVP