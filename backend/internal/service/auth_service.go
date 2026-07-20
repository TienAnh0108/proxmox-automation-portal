package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/dto"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

// AuthService định nghĩa các use-case liên quan tới xác thực.
// Handler sẽ phụ thuộc vào interface này, không phụ thuộc thẳng vào
// implementation cụ thể — cho phép mock khi viết unit test cho Handler.
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error)
	Logout(ctx context.Context, req dto.LogoutRequest) error
	ValidateAccessToken(tokenString string) (*Claims, error)
}

type authService struct {
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	tokenMgr    *TokenManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	tokenMgr *TokenManager,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		tokenMgr:    tokenMgr,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	role := domain.Role(req.Role)
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, postgres.ErrUsernameTaken) {
			return nil, err // Handler sẽ nhận diện lỗi này để trả 409 Conflict
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			// Cố ý trả lỗi CHUNG "invalid credentials" thay vì "user not found" —
			// tránh lộ thông tin username nào tồn tại trong hệ thống
			// (user enumeration attack).
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, hashedRefresh, err := s.tokenMgr.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshRecord := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenMgr.RefreshTokenTTL()),
	}
	if err := s.refreshRepo.Create(ctx, refreshRecord); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    s.tokenMgr.AccessTokenTTLSeconds(),
		User:         toUserResponse(user),
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error) {
	hashedToken := HashToken(req.RefreshToken)

	stored, err := s.refreshRepo.FindByTokenHash(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, postgres.ErrRefreshTokenNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	if stored.IsRevoked() {
		return nil, ErrTokenRevoked
	}
	if stored.IsExpired() {
		return nil, ErrTokenExpired
	}

	user, err := s.userRepo.FindById(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Rotation: thu hồi token cũ trước khi cấp token mới — nếu bước cấp mới
	// thất bại giữa chừng, token cũ vẫn bị vô hiệu (an toàn hơn để lộ khả năng
	// dùng lại token cũ so với việc tồn tại 2 token cùng hợp lệ).
	if err := s.refreshRepo.Revoke(ctx, stored.ID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, hashedRefresh, err := s.tokenMgr.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newRecord := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenMgr.RefreshTokenTTL()),
	}
	if err := s.refreshRepo.Create(ctx, newRecord); err != nil {
		return nil, fmt.Errorf("save new refresh token: %w", err)
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    s.tokenMgr.AccessTokenTTLSeconds(),
	}, nil
}

func (s *authService) Logout(ctx context.Context, req dto.LogoutRequest) error {
	hashedToken := HashToken(req.RefreshToken)

	stored, err := s.refreshRepo.FindByTokenHash(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, postgres.ErrRefreshTokenNotFound) {
			// Token không tồn tại — coi như đã logout, không cần báo lỗi
			// cho client (client chỉ quan tâm "giờ tôi đã đăng xuất chưa").
			return nil
		}
		return fmt.Errorf("find refresh token: %w", err)
	}

	return s.refreshRepo.Revoke(ctx, stored.ID)
}

func (s *authService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.tokenMgr.ParseAccessToken(tokenString)
}

func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
}
