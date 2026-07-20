package dto

import "time"

// RegisterRequest là dữ liệu admin gửi lên để tạo tài khoản mới.
// binding tag giúp Gin tự động validate trước khi vào tới Handler,
// tránh phải viết if kiểm tra thủ công cho từng field.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// LoginRequest là dữ liệu client gửi lên khi đăng nhập.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest là dữ liệu client gửi lên để lấy access token mới.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest là dữ liệu client gửi lên để thu hồi refresh token hiện tại.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse là dữ liệu User trả ra client — CHỦ Ý không có PasswordHash.
// Đây chính là lý do DTO tồn tại tách biệt với domain.User: nếu lỡ tay
// json.Marshal thẳng domain.User, PasswordHash sẽ bị lộ ra API response.
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse trả về sau khi đăng nhập thành công.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // đơn vị: giây, giúp client tự tính thời điểm hết hạn
	User         UserResponse `json:"user"`
}

// RefreshResponse trả về sau khi refresh thành công.
// Trả cả access_token lẫn refresh_token mới vì áp dụng cơ chế rotation
// (refresh token cũ bị revoke, cấp token mới thay thế).
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
