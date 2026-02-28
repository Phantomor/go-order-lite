package config

import "github.com/spf13/viper"

type Config struct {
	Server ServerConfig
	Log    LogConfig
	Mysql  MysqlConfig
}

type ServerConfig struct {
	Port int
}
type LogConfig struct {
	Level string
}
type MysqlConfig struct {
	DSN string
}

var Cfg Config

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(&Cfg)
}
