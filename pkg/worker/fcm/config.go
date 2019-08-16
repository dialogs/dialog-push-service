package fcm

import (
	"os"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`

	// Path to service account:
	// https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
	ServiceAccount string        `mapstructure:"service-account"`
	Retries        int           `mapstructure:"retries"`
	Timeout        time.Duration `mapstructure:"timeout"`
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

	if _, err := os.Stat(c.ServiceAccount); err != nil {
		return nil, errors.Wrap(err, "path to service-account")
	}

	return c, nil
}
