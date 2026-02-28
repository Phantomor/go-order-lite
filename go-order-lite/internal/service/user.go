package service

import (
	"go-order-lite/internal/dao"
	"go-order-lite/internal/model"
	"go-order-lite/pkg/errno"

	"golang.org/x/crypto/bcrypt"
)

func Register(username, password string) error {
	user, err := dao.GetUserByUsername(username)
	if err != nil {
		return err // 系统错误
	}
	if user != nil {
		return errno.UserExists // 业务错误
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	return dao.CreateUser(&model.User{
		Username: username,
		Password: string(hash),
	})
}
