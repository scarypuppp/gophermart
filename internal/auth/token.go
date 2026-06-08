package auth

type TokenPayload struct {
	UserID int64
}

type TokenProvider interface {
	CreateToken(userID int64) (string, error)
	ParseToken(tokenString string) (*TokenPayload, error)
}
