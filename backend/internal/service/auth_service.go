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

// AuthService — Login/Refresh trả thêm rawRefreshToken (string) tách biệt
// khỏi response JSON, vì Handler cần giá trị này để set cookie chứ không
// đưa vào JSON body. Logout nhận thẳng rawRefreshToken thay vì dto.LogoutRequest.
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, string, error)
	Refresh(ctx context.Context, rawRefreshToken string) (*dto.RefreshResponse, string, error)
	Logout(ctx context.Context, rawRefreshToken string) error
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
			return nil, err
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, string, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	rawRefresh, hashedRefresh, err := s.tokenMgr.GenerateRefreshToken()
	if err != nil {
		return nil, "", err
	}

	refreshRecord := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenMgr.RefreshTokenTTL()),
	}
	if err := s.refreshRepo.Create(ctx, refreshRecord); err != nil {
		return nil, "", fmt.Errorf("save refresh token: %w", err)
	}

	resp := &dto.LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   s.tokenMgr.AccessTokenTTLSeconds(),
		User:        toUserResponse(user),
	}
	return resp, rawRefresh, nil
}

func (s *authService) Refresh(ctx context.Context, rawRefreshToken string) (*dto.RefreshResponse, string, error) {
	hashedToken := HashToken(rawRefreshToken)

	stored, err := s.refreshRepo.FindByTokenHash(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, postgres.ErrRefreshTokenNotFound) {
			return nil, "", ErrInvalidToken
		}
		return nil, "", fmt.Errorf("find refresh token: %w", err)
	}

	if stored.IsRevoked() {
		return nil, "", ErrTokenRevoked
	}
	if stored.IsExpired() {
		return nil, "", ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("find user: %w", err)
	}

	if err := s.refreshRepo.Revoke(ctx, stored.ID); err != nil {
		return nil, "", fmt.Errorf("revoke old refresh token: %w", err)
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	rawRefresh, hashedRefresh, err := s.tokenMgr.GenerateRefreshToken()
	if err != nil {
		return nil, "", err
	}

	newRecord := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenMgr.RefreshTokenTTL()),
	}
	if err := s.refreshRepo.Create(ctx, newRecord); err != nil {
		return nil, "", fmt.Errorf("save new refresh token: %w", err)
	}

	resp := &dto.RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   s.tokenMgr.AccessTokenTTLSeconds(),
	}
	return resp, rawRefresh, nil
}

func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashedToken := HashToken(rawRefreshToken)

	stored, err := s.refreshRepo.FindByTokenHash(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, postgres.ErrRefreshTokenNotFound) {
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
