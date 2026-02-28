package dao

import (
	"errors"
	"go-order-lite/internal/model"
	"go-order-lite/pkg/mysql"

	"gorm.io/gorm"
)

func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := mysql.DB.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func CreateUser(user *model.User) error {
	return mysql.DB.Create(user).Error
}

func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := mysql.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
