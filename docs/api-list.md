# API List - Cloud Automation Portal (Proxmox)

Base URL: `http://10.8.5.248:8080`

## Nodes

### Lấy danh sách Node
- **Method:** GET
- **Endpoint:** `/api/nodes`
- **Mô tả:** Lấy danh sách tất cả node trong cluster Proxmox
- **Response mẫu:**
```json
[
  {
    "node": "proxmox",
    "status": "online",
    "cpu": 0.05,
    "maxmem": 16777216000,
    "mem": 8388608000
  }
]
```

---

## Virtual Machines (VM)

### Lấy danh sách VM theo Node
- **Method:** GET
- **Endpoint:** `/api/nodes/:node/vms`
- **Ví dụ:** `/api/nodes/proxmox/vms`
- **Response mẫu:**
```json
[
  {
    "vmid": 104,
    "name": "LabFrontendDeployment",
    "status": "running",
    "cpu": 0.02,
    "mem": 1073741824,
    "maxmem": 2147483648
  }
]
```

### Tạo mới VM
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms`
- **Ví dụ:** `/api/nodes/proxmox/vms`
- **Body (raw/JSON):**
```json
{
  "vmid": 9998,
  "name": "test-create-vm",
  "cores": 1,
  "memory": 512,
  "ostype": "l26"
}
```
- **Response mẫu:**
```json
{
  "message": "Đang tạo VM",
  "task_id": "UPID:proxmox:..."
}
```

### Xóa VM
- **Method:** DELETE
- **Endpoint:** `/api/nodes/:node/vms/:vmid`
- **Ví dụ:** `/api/nodes/proxmox/vms/9998`
- **Response mẫu:**
```json
{
  "message": "Đang xóa VM",
  "task_id": "UPID:proxmox:..."
}
```

### Start VM
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms/:vmid/start`
- **Ví dụ:** `/api/nodes/proxmox/vms/110/start`

### Stop VM (tắt cứng, giống rút nguồn)
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms/:vmid/stop`
- **Ví dụ:** `/api/nodes/proxmox/vms/110/stop`

### Shutdown VM (tắt đúng cách qua ACPI)
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms/:vmid/shutdown`
- **Ví dụ:** `/api/nodes/proxmox/vms/110/shutdown`

### Reboot VM (khởi động lại đúng cách)
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms/:vmid/reboot`
- **Ví dụ:** `/api/nodes/proxmox/vms/110/reboot`

### Reset VM (khởi động lại kiểu cứng)
- **Method:** POST
- **Endpoint:** `/api/nodes/:node/vms/:vmid/reset`
- **Ví dụ:** `/api/nodes/proxmox/vms/110/reset`

- **Response mẫu (chung cho start/stop/shutdown/reboot/reset):**
```json
{
  "message": "Lệnh đã được gửi",
  "task_id": "UPID:proxmox:..."
}
```

---

## Ghi chú kỹ thuật

- Xác thực: Proxmox API Token (`PVEAPIToken=<tokenID>=<tokenSecret>`)
- Server chạy trên VM nội bộ công ty (`10.8.5.248:8080`), kết nối tới Proxmox qua IP nội bộ `10.8.5.249:8006`
- Cấu hình đọc từ `backend/configs/config.env` (không commit lên Git, có trong `.gitignore`)
- Framework: Gin (Go)
- HTTP client: Resty (`SetContentLength(true)` bắt buộc để tránh lỗi "chunked transfer encoding not supported" từ Proxmox khi gửi request không có body)