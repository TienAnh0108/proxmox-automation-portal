package repository

import (
	"context"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	FindByUPID(ctx context.Context, upid string) (*domain.Task, error)
	UpdateStatus(ctx context.Context, upid string, status domain.TaskStatus, exitStatus *string) error
}
