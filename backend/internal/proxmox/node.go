package proxmox

import "fmt"

// rawNode dùng để nhận dữ liệu byte thô từ Proxmox API (không public ra ngoài)
type rawNode struct {
	Node    string  `json:"node"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	MaxCPU  int     `json:"maxcpu"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	Disk    int64   `json:"disk"`
	MaxDisk int64   `json:"maxdisk"`
}

// Node là dữ liệu trả về cho client - chỉ chứa GiB và %, không có byte thô
type Node struct {
	Node        string  `json:"node"`
	Status      string  `json:"status"`
	CPUPercent  float64 `json:"cpu_percent"`
	MaxCPU      int     `json:"maxcpu"`
	MemGiB      float64 `json:"mem_gib"`
	MaxMemGiB   float64 `json:"maxmem_gib"`
	MemPercent  float64 `json:"mem_percent"`
	DiskGiB     float64 `json:"disk_gib"`
	MaxDiskGiB  float64 `json:"maxdisk_gib"`
	DiskPercent float64 `json:"disk_percent"`
}

type nodeResponse struct {
	Data []rawNode `json:"data"`
}

func (c *Client) ListNodes() ([]Node, error) {
	var result nodeResponse

	resp, err := c.client.R().
		SetResult(&result).
		Get("/nodes")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Proxmox API error: %s", resp.String())
	}

	nodes := make([]Node, 0, len(result.Data))
	for _, raw := range result.Data {
		nodes = append(nodes, Node{
			Node:        raw.Node,
			Status:      raw.Status,
			CPUPercent:  roundToPercent(raw.CPU),
			MaxCPU:      raw.MaxCPU,
			MemGiB:      bytesToGiB(raw.Mem),
			MaxMemGiB:   bytesToGiB(raw.MaxMem),
			MemPercent:  calcPercent(raw.Mem, raw.MaxMem),
			DiskGiB:     bytesToGiB(raw.Disk),
			MaxDiskGiB:  bytesToGiB(raw.MaxDisk),
			DiskPercent: calcPercent(raw.Disk, raw.MaxDisk),
		})
	}

	return nodes, nil
}
