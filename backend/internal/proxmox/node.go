package proxmox

import "fmt"

type Node struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	MaxMem int64   `json:"maxmem"`
	Mem    int64   `json:"mem"`
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
		return nil, fmt.Errorf("proxmox API error: %s", resp.String())
	}

	return result.Data, nil
}
