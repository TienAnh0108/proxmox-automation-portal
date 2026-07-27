package proxmox

import (
	"fmt"
	"strconv"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/dto"
)

// rawVM nhận dữ liệu byte thô từ Proxmox
type rawVM struct {
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	Mem      int64   `json:"mem"`
	MaxMem   int64   `json:"maxmem"`
	MaxDisk  int64   `json:"maxdisk"`
	Template int     `json:"template"`
}

// Chứa thêm uptime, cores, sockets fileds hơn là rawVM
type rawVMStatus struct {
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	CPUs     int     `json:"cpus"`
	Mem      int64   `json:"mem"`
	MaxMem   int64   `json:"maxmem"`
	MaxDisk  int64   `json:"maxdisk"`
	Uptime   int64   `json:"uptime"`
	Template int     `json:"template"`
}

// VMDetail là dữ liệu chi tiết trả cho client khi xem 1 VM cụ thể.
type VMDetail struct {
	VMID          int     `json:"vmid"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	CPUPercent    float64 `json:"cpu_percent"`
	Cores         int     `json:"cores"`
	MemGiB        float64 `json:"mem_gib"`
	MaxMemGiB     float64 `json:"maxmem_gib"`
	MemPercent    float64 `json:"mem_percent"`
	MaxDiskGiB    float64 `json:"maxdisk_gib"`
	IsTemplate    bool    `json:"is_template"`
}

// VM là dữ liệu trả về cho client - chỉ chứa GiB và %
type VM struct {
	VMID       int     `json:"vmid"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpu_percent"`
	MemGiB     float64 `json:"mem_gib"`
	MaxMemGiB  float64 `json:"maxmem_gib"`
	MemPercent float64 `json:"mem_percent"`
	MaxDiskGiB float64 `json:"maxdisk_gib"`
	IsTemplate bool    `json:is_template`
}

type vmStatusResponse struct {
	Data rawVMStatus `json:"data"`
}

type vmResponse struct {
	Data []rawVM `json:"data"`
}

type taskResponse struct {
	Data string `json:"data"`
}

func (c *Client) ListVMs(node string) ([]VM, error) {
	var result vmResponse

	resp, err := c.client.R().
		SetResult(&result).
		Get(fmt.Sprintf("/nodes/%s/qemu", node))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("proxmox API error: %s", resp.String())
	}

	vms := make([]VM, 0, len(result.Data))
	for _, raw := range result.Data {
		vms = append(vms, VM{
			VMID:       raw.VMID,
			Name:       raw.Name,
			Status:     raw.Status,
			CPUPercent: roundToPercent(raw.CPU),
			MemGiB:     bytesToGiB(raw.Mem),
			MaxMemGiB:  bytesToGiB(raw.MaxMem),
			MemPercent: calcPercent(raw.Mem, raw.MaxMem),
			MaxDiskGiB: bytesToGiB(raw.MaxDisk),
			IsTemplate: raw.Template == 1,
		})
	}

	return vms, nil
}

// --- Các hàm điều khiển VM giữ nguyên như cũ, không đổi ---

func (c *Client) vmAction(node string, vmid int, action string) (string, error) {
	var result taskResponse

	resp, err := c.client.R().
		SetResult(&result).
		Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/%s", node, vmid, action))

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("proxmox API error: %s", resp.String())
	}

	return result.Data, nil
}

// GetVMDetail lấy thông tin chi tiết 1 VM — dùng cho tính năng "Xem chi tiết VM".
func (c *Client) GetVMDetail(node string, vmid int) (*VMDetail, error) {
	var result vmStatusResponse

	resp, err := c.client.R().
		SetResult(&result).
		Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", node, vmid))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("proxmox API error: %s", resp.String())
	}

	raw := result.Data
	return &VMDetail{
		VMID:          raw.VMID,
		Name:          raw.Name,
		Status:        raw.Status,
		IsTemplate:    raw.Template == 1,
		UptimeSeconds: raw.Uptime,
		CPUPercent:    roundToPercent(raw.CPU),
		Cores:         raw.CPUs,
		MemGiB:        bytesToGiB(raw.Mem),
		MaxMemGiB:     bytesToGiB(raw.MaxMem),
		MemPercent:    calcPercent(raw.Mem, raw.MaxMem),
		MaxDiskGiB:    bytesToGiB(raw.MaxDisk),
	}, nil
}

func (c *Client) StartVM(node string, vmid int) (string, error) {
	return c.vmAction(node, vmid, "start")
}

func (c *Client) StopVM(node string, vmid int) (string, error) {
	return c.vmAction(node, vmid, "stop")
}

func (c *Client) ShutdownVM(node string, vmid int) (string, error) {
	return c.vmAction(node, vmid, "shutdown")
}

func (c *Client) RebootVM(node string, vmid int) (string, error) {
	return c.vmAction(node, vmid, "reboot")
}

func (c *Client) ResetVM(node string, vmid int) (string, error) {
	return c.vmAction(node, vmid, "reset")
}

type CreateVMRequest struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Cores  int    `json:"cores"`
	Memory int    `json:"memory"`
	OSType string `json:"ostype"`
}

// Create Virtual Machine
func (c *Client) CloneVM(node string, templateVMID int, req dto.CloneVMRequest) (string, error) {
	var result taskResponse

	formData := map[string]string{
		"newid": strconv.Itoa(req.NewVMID),
		"name":  req.Name,
	}
	if req.FullClone {
		formData["full"] = "1"
	} else {
		formData["full"] = "0"
	}

	resp, err := c.client.R().
		SetFormData(formData).
		SetResult(&result).
		Post(fmt.Sprintf("/nodes/%s/qemu/%d/clone", node, templateVMID))

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("proxmox API error: %s", resp.String())
	}

	return result.Data, nil
}

// Delete Virtual Machine
func (c *Client) DeleteVM(node string, vmid int) (string, error) {
	var result taskResponse

	resp, err := c.client.R().
		SetResult(&result).
		Delete(fmt.Sprintf("/nodes/%s/qemu/%d", node, vmid))

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("proxmox API error: %s", resp.String())
	}

	return result.Data, nil
}
