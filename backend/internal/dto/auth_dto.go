package dto

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest và LogoutRequest KHÔNG còn cần thiết — refresh token giờ
// đọc từ cookie HttpOnly, không còn nằm trong request body nữa.

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse KHÔNG còn field RefreshToken — refresh token được set qua
// Set-Cookie header (HttpOnly), không lộ ra JSON body để JavaScript không
// đọc được (chống XSS đánh cắp token).
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// RefreshResponse tương tự — không còn RefreshToken trong body.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
