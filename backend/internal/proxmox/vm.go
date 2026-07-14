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

type taskResponse struct {
	Data string `json:"data"`
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

// vmAction gửi lệnh điều khiển VM (start/stop/shutdown/reboot/reset)
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

	return result.Data, nil // Trả về UPID để theo dõi task
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
