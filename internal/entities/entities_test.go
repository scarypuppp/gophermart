package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogin(t *testing.T) {
	assert.NoError(t, ValidateLogin("alice"))
	assert.ErrorIs(t, ValidateLogin("ab"), ErrIncorrectLoginLength)
	assert.ErrorIs(t, ValidateLogin(""), ErrIncorrectLoginLength)
}

func TestValidatePassword(t *testing.T) {
	assert.NoError(t, ValidatePassword("secret"))
	assert.ErrorIs(t, ValidatePassword("abc"), ErrIncorrectPasswordLength)
	assert.ErrorIs(t, ValidatePassword(""), ErrIncorrectPasswordLength)
}

func TestValidateOrderNumber(t *testing.T) {
	assert.True(t, ValidateOrderNumber("12345678903"))
	assert.False(t, ValidateOrderNumber("12345678900"))
	assert.False(t, ValidateOrderNumber(""))
	assert.False(t, ValidateOrderNumber("abc"))
}
