package secutils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
}

func ComparePassword(hashed []byte, pwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashed, []byte(pwd))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
