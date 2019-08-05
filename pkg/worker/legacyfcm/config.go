package legacyfcm

import (
	"strings"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2legacyfcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`
	APIConfig      *api2legacyfcm.Config `mapstructure:"-"`

	// Server key:
	// https://console.firebase.google.com/project/_/settings/cloudmessaging/
	ServerKey string `mapstructure:"key"`
	SendTries int    `mapstructure:"retries"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	c := &Config{}
	err := src.Unmarshal(c)
	if err != nil {
		return nil, err
	}

	c.Config, err = worker.NewConfig(src)
	if err != nil {
		return nil, err
	}

	switch c.ConverterKind {
	case converter.KindApi:
		c.APIConfig, err = api2legacyfcm.NewConfig(src)
	case converter.KindBinary:
		// nothing do
	default:
		err = errors.New("invalid converter config kind")
	}

	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(c.ServerKey) == "" {
		return nil, errors.New("invalid server key")
	}

	return c, nil
}
