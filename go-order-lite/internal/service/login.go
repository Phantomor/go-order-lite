package service

import (
	"go-order-lite/internal/dao"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

func Login(username string, password string) (string, error) {
	user, err := dao.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errno.UserNotFound
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	); err != nil {
		return "", errno.PasswordWrong
	}
	return jwt.GenerateToken(user.ID)
}
