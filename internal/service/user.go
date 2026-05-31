package service

import (
	"context"
	"errors"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/utils/auth"
)

//
//	ERRORS
//

var ErrLoginAlreadyExists = errors.New("user with such login already exists")
var ErrLoginPasswordNotExist = errors.New("user with such login and password does not exist")

//
//	INTERFACES
//

type UserRepository interface {
	GetById(ctx context.Context, id int64) (*entities.User, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	CreateUser(ctx context.Context, user entities.User) (*entities.User, error)
}

//
//	SERVICE
//

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return UserService{repo}
}

// RegisterUser создает пользователя по логину и паролю.
func (us *UserService) RegisterUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {

	err := entities.ValidatePassword(password)
	if err != nil {
		return nil, err
	}
	err = entities.ValidateLogin(login)
	if err != nil {
		return nil, err
	}

	existingUser, err := us.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrLoginAlreadyExists
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := us.repo.CreateUser(ctx, entities.User{Login: login, Password: passwordHash})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// LoginUser проверяет, существует ли такой пользователь по предоставленным логину и паролю
func (us *UserService) LoginUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {
	user, err := us.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrLoginPasswordNotExist
	}
	err = auth.CheckPassword(password, user.Password)
	if err != nil {
		return nil, ErrLoginPasswordNotExist
	}
	return user, err
}
