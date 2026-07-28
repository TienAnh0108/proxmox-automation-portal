package dto

import "time"

type TaskResponse struct {
	UPID       string    `json:"upid"`
	Node       string    `json:"node"`
	VMID       int       `json:"vmid"`
	Action     string    `json:"action"`
	Status     string    `json:"status"`
	ExitStatus *string   `json:"exit_status,omitempty"`
	Success    *bool     `json:"success,omitempty"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}
