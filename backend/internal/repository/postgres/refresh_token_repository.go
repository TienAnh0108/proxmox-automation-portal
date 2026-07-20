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

// ErrRefreshTokenNotFound báo không tìm thấy refresh token trong DB —
// dùng khi token không tồn tại, hoặc client gửi token giả mạo.
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

type refreshTokenRow struct {
	ID        string       `db:"id"`
	UserID    string       `db:"user_id"`
	TokenHash string       `db:"token_hash"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
	CreatedAt time.Time    `db:"created_at"`
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	token.ID = uuid.NewString()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	err := r.db.QueryRowxContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var row refreshTokenRow
	query := `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
	          FROM refresh_tokens WHERE token_hash = $1`

	err := r.db.GetContext(ctx, &row, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	return refreshTokenRowToDomain(row), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	// Kiểm tra có thực sự update dòng nào không — nếu id không tồn tại,
	// UPDATE vẫn "thành công" (không lỗi) nhưng không đổi gì cả.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func refreshTokenRowToDomain(row refreshTokenRow) *domain.RefreshToken {
	rt := &domain.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		rt.RevokedAt = &t
	}
	return rt
}
