package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

type taskRow struct {
	ID            string         `db:"id"`
	UPID          string         `db:"upid"`
	Node          string         `db:"node"`
	VMID          int            `db:"vmid"`
	Action        string         `db:"action"`
	Status        string         `db:"status"`
	ExitStatus    sql.NullString `db:"exit_status"`
	CreatedBy     string         `db:"created_by"`
	CreatedAt     time.Time      `db:"created_at"`
	LastCheckedAt time.Time      `db:"last_checked_at"`
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	task.ID = uuid.NewString()

	query := `
		INSERT INTO tasks (id, upid, node, vmid, action, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, last_checked_at`

	err := r.db.QueryRowxContext(ctx, query,
		task.ID, task.UPID, task.Node, task.VMID, task.Action,
		string(task.Status), task.CreatedBy,
	).Scan(&task.CreatedAt, &task.LastCheckedAt)

	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepository) FindByUPID(ctx context.Context, upid string) (*domain.Task, error) {
	var row taskRow
	query := `SELECT id, upid, node, vmid, action, status, exit_status,
	                 created_by, created_at, last_checked_at
	          FROM tasks WHERE upid = $1`

	err := r.db.GetContext(ctx, &row, query, upid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task by upid: %w", err)
	}
	return taskRowToDomain(row), nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, upid string, status domain.TaskStatus, exitStatus *string) error {
	query := `
		UPDATE tasks
		SET status = $1, exit_status = $2, last_checked_at = now()
		WHERE upid = $3`

	result, err := r.db.ExecContext(ctx, query, string(status), exitStatus, upid)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func taskRowToDomain(row taskRow) *domain.Task {
	t := &domain.Task{
		ID:            row.ID,
		UPID:          row.UPID,
		Node:          row.Node,
		VMID:          row.VMID,
		Action:        row.Action,
		Status:        domain.TaskStatus(row.Status),
		CreatedBy:     row.CreatedBy,
		CreatedAt:     row.CreatedAt,
		LastCheckedAt: row.LastCheckedAt,
	}
	if row.ExitStatus.Valid {
		s := row.ExitStatus.String
		t.ExitStatus = &s
	}
	return t
}
