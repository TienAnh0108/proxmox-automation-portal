package proxmox

import "fmt"

type TaskStatusResult struct {
	Status     string // "running" | "stopped"
	ExitStatus string // rỗng nếu chưa xong, "OK" nếu thành công, mô tả lỗi nếu fail
}

type taskStatusRaw struct {
	Status     string `json:"status"`
	ExitStatus string `json:"exitstatus"`
}

type taskStatusResponse struct {
	Data taskStatusRaw `json:"data"`
}

// GetTaskStatus poll trạng thái mới nhất của 1 task theo UPID từ Proxmox.
func (c *Client) GetTaskStatus(node, upid string) (*TaskStatusResult, error) {
	var result taskStatusResponse

	resp, err := c.client.R().
		SetResult(&result).
		Get(fmt.Sprintf("/nodes/%s/tasks/%s/status", node, upid))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("proxmox API error: %s", resp.String())
	}

	return &TaskStatusResult{
		Status:     result.Data.Status,
		ExitStatus: result.Data.ExitStatus,
	}, nil
}
