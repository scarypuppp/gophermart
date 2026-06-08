package auth

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password string, hash string) error
}
