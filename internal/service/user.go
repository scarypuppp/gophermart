package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophermart/internal/auth"
	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/repository"
)

//
//	ERRORS
//

var ErrLoginAlreadyExists = errors.New("user with such login already exists")
var ErrLoginPasswordNotExist = errors.New("user with such login and password does not exist")

//
//	SERVICE
//

type UserService struct {
	uow            repository.UnitOfWork
	passwordHasher auth.PasswordHasher
}

func NewUserService(uow repository.UnitOfWork, passwordHasher auth.PasswordHasher) *UserService {
	return &UserService{uow, passwordHasher}
}

// RegisterUser создает пользователя по логину и паролю.
func (us *UserService) RegisterUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {

	err := entities.ValidateLogin(login)
	if err != nil {
		return nil, err
	}
	err = entities.ValidatePassword(password)
	if err != nil {
		return nil, err
	}

	tx, err := us.uow.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Users().GetByLogin(ctx, login)
	if err == nil {
		return nil, ErrLoginAlreadyExists
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, fmt.Errorf("RegisterUser: check login: %w", err)
	}

	passwordHash, err := us.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: hash password: %w", err)
	}

	user, err := tx.Users().CreateUser(ctx, entities.User{Login: login, Password: passwordHash})
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: create user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("RegisterUser: commit: %w", err)
	}
	return user, nil
}

// LoginUser проверяет, существует ли такой пользователь по предоставленным логину и паролю
func (us *UserService) LoginUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {
	user, err := us.uow.Users().GetByLogin(ctx, login)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrLoginPasswordNotExist
	}
	if err != nil {
		return nil, fmt.Errorf("LoginUser: %w", err)
	}
	err = us.passwordHasher.CheckPassword(password, user.Password)
	if err != nil {
		return nil, ErrLoginPasswordNotExist
	}

	return user, nil
}
