package api2legacyfcm

import (
	"github.com/spf13/viper"
)

type Config struct {
	AllowAlerts bool `mapstructure:"allow-alerts"`
	Sandbox     bool `mapstructure:"sandbox"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	c := &Config{}
	if err := src.Unmarshal(c); err != nil {
		return nil, err
	}

	return c, nil
}
