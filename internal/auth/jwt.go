package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TOKEN_EXP = time.Hour * 24 * 30 * 6

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func CreateToken(secretKey string, userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(secretKey string, tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return -1, err
	}
	if !token.Valid {
		return -1, fmt.Errorf("%w", token.Claims.Valid())
	}
	return claims.UserID, nil
}
