package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword Функция для генерации хэша от пароля
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword Функция для проверки корректности введенного пароля
func CheckPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
