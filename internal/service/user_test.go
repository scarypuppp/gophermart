package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/scarypuppp/gophermart/internal/entities"
	"github.com/scarypuppp/gophermart/internal/utils/auth"
	"github.com/stretchr/testify/assert"
)

type MockUserRepository struct {
	currentUserId atomic.Int64
	users         []entities.User

	mu sync.RWMutex
}

func NewMockRepository(users []entities.User) *MockUserRepository {
	return &MockUserRepository{users: users}
}

func (m *MockUserRepository) GetById(ctx context.Context, id int64) (*entities.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.ID == id {
			uCopy := u
			return &uCopy, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Login == login {
			uCopy := u
			return &uCopy, nil
		}
	}
	return nil, nil
}
func (m *MockUserRepository) CreateUser(ctx context.Context, user entities.User) (*entities.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUserId.Add(1)
	user.ID = m.currentUserId.Load()
	m.users = append(m.users, user)
	return &user, nil
}

func TestRegisterUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		us := NewUserService(NewMockRepository([]entities.User{}))

		user, err := us.RegisterUser(context.TODO(), "user", "12345")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user", user.Login)
		assert.NotZero(t, user.ID)
		assert.NotEqual(t, "12345", user.Password)
	})

	t.Run("incorrect login", func(t *testing.T) {
		us := NewUserService(NewMockRepository([]entities.User{}))

		_, err := us.RegisterUser(context.TODO(), "", "12345")

		assert.Error(t, err)
		assert.ErrorIs(t, err, entities.ErrIncorrectLoginLength)
	})
	t.Run("incorrect password", func(t *testing.T) {
		us := NewUserService(NewMockRepository([]entities.User{}))

		_, err := us.RegisterUser(context.TODO(), "user", "1")

		assert.Error(t, err)
		assert.ErrorIs(t, err, entities.ErrIncorrectPasswordLength)
	})

	t.Run("duplicate login", func(t *testing.T) {
		existing := []entities.User{{ID: 1, Login: "user"}}
		us := NewUserService(NewMockRepository(existing))

		_, err := us.RegisterUser(context.TODO(), "user", "12345")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrLoginAlreadyExists)
	})
}

func TestLoginUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var users []entities.User
		passwordHash, _ := auth.HashPassword("12345")
		users = append(users, entities.User{ID: 1, Login: "user", Password: passwordHash})
		us := NewUserService(NewMockRepository(users))

		user, err := us.LoginUser(context.TODO(), "user", "12345")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.Login, users[0].Login)
		assert.Equal(t, user.ID, users[0].ID)
	})
	t.Run("user not exist", func(t *testing.T) {
		var users []entities.User
		passwordHash, _ := auth.HashPassword("12345")
		users = append(users, entities.User{ID: 1, Login: "user", Password: passwordHash})
		us := NewUserService(NewMockRepository(users))

		user, err := us.LoginUser(context.TODO(), "user2", "12345")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrLoginPasswordNotExist)
		assert.Nil(t, user)
	})
	t.Run("incorrect password", func(t *testing.T) {
		var users []entities.User
		passwordHash, _ := auth.HashPassword("12345")
		users = append(users, entities.User{ID: 1, Login: "user", Password: passwordHash})
		us := NewUserService(NewMockRepository(users))

		user, err := us.LoginUser(context.TODO(), "user", "54321")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrLoginPasswordNotExist)
		assert.Nil(t, user)
	})
}
