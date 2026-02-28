package mysql

import (
	"go-order-lite/internal/model"
	"go-order-lite/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	db, err := gorm.Open(mysql.Open(config.Cfg.Mysql.DSN), &gorm.Config{})
	if err != nil {
		return err
	}
	// 自动建表
	if err := db.AutoMigrate(&model.User{}, &model.Order{}); err != nil {
		return err
	}

	DB = db
	return nil
}
