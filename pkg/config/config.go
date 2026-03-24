package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Mysql    MysqlConfig
	RocketMQ RocketMQConfig
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

type RocketMQConfig struct {
	NameServers []string `mapstructure:"name_servers"`
	Retry       int      `mapstructure:"retry"`
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
