package fcm

import (
	"os"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`
	APIConfig      *api2fcm.Config `mapstructure:"-"`

	// Path to service account:
	// https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
	ServiceAccount string        `mapstructure:"service-account"`
	SendTries      int           `mapstructure:"send-tries"`
	SendTimeout    time.Duration `mapstructure:"send-timeout"`
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
		c.APIConfig, err = api2fcm.NewConfig(src)
	case converter.KindBinary:
		// nothing do
	default:
		err = errors.New("invalid converter config kind")
	}

	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(c.ServiceAccount); err != nil {
		return nil, errors.Wrap(err, "path to service-account")
	}

	return c, nil
}
