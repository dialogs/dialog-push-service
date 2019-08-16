package worker

import (
	"github.com/dialogs/dialog-push-service/pkg/conversion"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*conversion.Config `mapstructure:"-"`
	ProjectID          string `mapstructure:"project-id"`
	NopMode            bool   `mapstructure:"nop-mode"`
	CountThreads       int    `mapstructure:"workers"`
	Sandbox            bool   `mapstructure:"sandbox"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	conversionConfig := &conversion.Config{}
	if err := src.Unmarshal(conversionConfig); err != nil {
		return nil, err
	}

	c := &Config{}
	if err := src.Unmarshal(c); err != nil {
		return nil, err
	}

	c.Config = conversionConfig
	if len(c.ProjectID) == 0 {
		return nil, errors.New("invalid `project-id`")
	}

	return c, nil
}
