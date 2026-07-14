package proxmox

import "fmt"

type Node struct {
	Node        string  `json:"node"`
	Status      string  `json:"status"`
	CPU         float64 `json:"cpu"`
	CPUPercent  float64 `json:"cpu_percent"`
	MaxCPU      int     `json:"maxcpu`
	MaxMem      int64   `json:"maxmem"`
	Mem         int64   `json:"mem"`
	MemPercent  float64 `json:"mem_percent"`
	Disk        int64   `json:"disk"`
	MaxDisk     int64   `json:"maxdisk`
	DiskPercent float64 `json:"disk_percent"`
}

type nodeResponse struct {
	Data []Node `json:"data"`
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

	for i := range result.Data {
		result.Data[i].CPUPercent = roundToPercent(result.Data[i].CPU)
		result.Data[i].MemPercent = calcPercent(result.Data[i].Mem, result.Data[i].MaxMem)
		result.Data[i].DiskPercent = calcPercent(result.Data[i].Disk, result.Data[i].MaxDisk)
	}

	return result.Data, nil
}
