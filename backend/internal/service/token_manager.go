package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"usernam"`
	Role     domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateAccessToken ký 1 JWT chứa thông tin user, hết hạn sau accessTokenTTL.
func (m *TokenManager) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken verify chữ ký + hạn JWT, trả về Claims nếu hợp lệ.
// Middleware sẽ gọi hàm này ở mọi request tới route được bảo vệ.
func (m *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Kiểm tra thuật toán ký đúng như mong đợi — chặn kiểu tấn công
		// "algorithm confusion" khi kẻ tấn công đổi alg trong header JWT
		// sang "none" hoặc thuật toán khác để bypass verify.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken sinh 1 chuỗi ngẫu nhiên (KHÔNG phải JWT) làm refresh token.
// Trả về 2 giá trị: raw token (gửi cho client) và hash của nó (lưu vào DB).
// Không bao giờ lưu raw token vào DB — nếu DB bị lộ, kẻ tấn công không thể
// dùng hash để giả mạo token (một chiều, không đảo ngược được).
func (m *TokenManager) GenerateRefreshToken() (raw string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}

	raw = base64.URLEncoding.EncodeToString(bytes)
	hash = HashToken(raw)
	return raw, hash, nil
}

// HashToken băm refresh token bằng SHA-256 — dùng SHA-256 (không phải bcrypt)
// vì đây không phải password do người dùng nghĩ ra (dễ đoán), mà là chuỗi
// ngẫu nhiên 256-bit — SHA-256 đủ an toàn, nhanh hơn bcrypt nhiều lần,
// phù hợp vì hàm này chạy trên MỌI request refresh.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (m *TokenManager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

func (m *TokenManager) AccessTokenTTLSeconds() int {
	return int(m.accessTokenTTL.Seconds())
}
