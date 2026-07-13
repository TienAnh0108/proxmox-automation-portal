package proxmox

import "fmt"

type VM struct {
	VMID   int     `json:"vmid"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

type vmResponse struct {
	Data []VM `json:"data"`
}

// ListVMs lấy danh sách VM trên 1 node cụ thể
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

	return result.Data, nil
}
