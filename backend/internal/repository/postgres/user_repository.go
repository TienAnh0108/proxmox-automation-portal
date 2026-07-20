package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ErrUserNotFound là sentinel error — Service layer sẽ so sánh với lỗi này
// bằng errors.Is() để quyết định trả 404 thay vì 500 cho client.
var ErrUserNotFound = errors.New("user not found")

// ErrUsernameTaken báo username đã tồn tại (vi phạm UNIQUE constraint).
var ErrUsernameTaken = errors.New("username already taken")

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// userRow ánh xạ trực tiếp với cột trong bảng `users` — tách riêng khỏi
// domain.User vì có thể khác biệt nhỏ (ví dụ tag `db` chỉ cần ở đây,
// domain layer không nên biết gì về cách lưu trữ).
type userRow struct {
	ID           string `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.NewString()

	query := `
		INSERT INTO users (id, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Username, user.PasswordHash, string(user.Role),
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Postgres trả mã lỗi 23505 khi vi phạm UNIQUE constraint.
		// Kiểm tra chuỗi lỗi thay vì so sánh code cụ thể để không phải
		// import thêm driver-specific error package — đơn giản hơn cho MVP.
		if isUniqueViolation(err) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var row userRow
	query := `SELECT id, username, password_hash, role, created_at, updated_at
	          FROM users WHERE username = $1`

	err := r.db.GetContext(ctx, &row, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return rowToDomain(row), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var row userRow
	query := `SELECT id, username, password_hash, role, created_at, updated_at
	          FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return rowToDomain(row), nil
}

func rowToDomain(row userRow) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Role:         domain.Role(row.Role),
	}
}
