package api2ans

import (
	"github.com/spf13/viper"
)

type Config struct {
	Topic       string `mapstructure:"topic"`
	AllowAlerts bool   `mapstructure:"allow-alerts"`
	Sound       string `mapstructure:"sound"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	c := &Config{}
	if err := src.Unmarshal(c); err != nil {
		return nil, err
	}

	return c, nil
}
