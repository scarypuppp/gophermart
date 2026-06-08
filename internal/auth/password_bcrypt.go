package auth

import "golang.org/x/crypto/bcrypt"

type PasswordHasherBcrypt struct{}

func NewPasswordHasherBcrypt() *PasswordHasherBcrypt {
	return &PasswordHasherBcrypt{}
}

// HashPassword Функция для генерации хэша от пароля
func (ph *PasswordHasherBcrypt) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword Функция для проверки корректности введенного пароля
func (ph *PasswordHasherBcrypt) CheckPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
