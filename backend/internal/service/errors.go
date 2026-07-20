package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("Invalid username and password")
	ErrInvalidRole        = errors.New("Invalid role")
	ErrTokenExpired       = errors.New("Refresh token expired")
	ErrTokenRevoked       = errors.New("Refresh token revoked")
	ErrInvalidToken       = errors.New("Invalid token")
)
