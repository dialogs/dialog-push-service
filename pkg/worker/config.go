package worker

import (
	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	ProjectID     string         `mapstructure:"project-id"`
	NopMode       bool           `mapstructure:"nop-mode"`
	CountThreads  int            `mapstructure:"workers"`
	ConverterKind converter.Kind `mapstructure:"-"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	c := &Config{}
	err := src.Unmarshal(c)
	if err != nil {
		return nil, err
	}

	if len(c.ProjectID) == 0 {
		return nil, errors.New("invalid `project-id`")
	}

	c.ConverterKind = converter.GetKindFromConfig(src)

	return c, nil
}
