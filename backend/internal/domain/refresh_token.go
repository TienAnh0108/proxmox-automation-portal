package domain

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time // con trỏ vì có thể NULL — chưa bị revoke
	CreatedAt time.Time
}

// IsExpired kiểm tra token đã hết hạn chưa — business rule thuộc về domain.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked kiểm tra token đã bị thu hồi chưa.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
