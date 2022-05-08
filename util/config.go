package util

import (
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 string        `mapstructure:"port"`
	DummyMessageSize     string        `mapstructure:"dummy_message_size"`
	DummyMessageDuration time.Duration `mapstructure:"dummy_message_duration"`
}

type ClientConfig struct {
	RequestDuration time.Duration `mapstructure:"request_duration"`
	TotalRequest    int           `mapstructure:"total_request"`
}

type Config struct {
	Client ClientConfig `mapstructure:"client"`
	Server ServerConfig `mapstructure:"server"`
}

func LoadConfig(filename string) (config Config, err error) {
	viper.AddConfigPath(".")

	viper.SetConfigName(filename)

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
