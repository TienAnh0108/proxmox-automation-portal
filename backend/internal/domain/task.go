package domain

import "time"

type TaskStatus string

const (
	TaskStatusRunning TaskStatus = "running"
	TaskStatusStopped TaskStatus = "stopped"
)

type Task struct {
	ID            string
	UPID          string
	Node          string
	VMID          int
	Action        string
	Status        TaskStatus
	ExitStatus    *string
	CreatedBy     string
	CreatedAt     time.Time
	LastCheckedAt time.Time
}

func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusRunning
}
