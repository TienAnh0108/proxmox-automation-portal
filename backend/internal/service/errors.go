package service

import (
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("Invalid username and password")
	ErrInvalidRole        = errors.New("Invalid role")
	ErrTokenExpired       = errors.New("Refresh token expired")
	ErrTokenRevoked       = errors.New("Refresh token revoked")
	ErrInvalidToken       = errors.New("Invalid token")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskForbidden      = errors.New("not allowed to view this task")
	ErrVMInvalidState     = errors.New("vm invalid stats for action")
)

type VMStateError struct {
	Message string
}

func (e *VMStateError) Error() string {
	return e.Message
}

func (e *VMStateError) Unwrap() error {
	return ErrVMInvalidState
}
