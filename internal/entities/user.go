package entities

import (
	"fmt"
)

const (
	MinLoginLength    = 3
	MaxLoginLength    = 64
	MinPasswordLength = 5
	MaxPasswordLength = 72
)

var (
	ErrIncorrectLoginLength    = fmt.Errorf("login length should be between %d and %d", MinLoginLength, MaxLoginLength)
	ErrIncorrectPasswordLength = fmt.Errorf("password length should be between %d and %d", MinPasswordLength, MaxPasswordLength)
)

type User struct {
	ID       int64  `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

func ValidateLogin(login string) error {
	if len(login) < MinLoginLength || len(login) > MaxLoginLength {
		return ErrIncorrectLoginLength
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return ErrIncorrectPasswordLength
	}
	return nil
}
