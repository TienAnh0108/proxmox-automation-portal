package proxmox

import (
	"fmt"
	"strconv"
)

// rawVM nhận dữ liệu byte thô từ Proxmox
type rawVM struct {
	VMID    int     `json:"vmid"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	MaxDisk int64   `json:"maxdisk"`
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
func (c *Client) CreateVM(node string, req CreateVMRequest) (string, error) {
	var result taskResponse

	resp, err := c.client.R().
		SetFormData(map[string]string{
			"vmid":   strconv.Itoa(req.VMID),
			"name":   req.Name,
			"cores":  strconv.Itoa(req.Cores),
			"memory": strconv.Itoa(req.Memory),
			"ostype": req.OSType,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("/nodes/%s/qemu", node))

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
