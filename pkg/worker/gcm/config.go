package gcm

import (
	"strings"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`

	// Server key:
	// https://console.firebase.google.com/project/_/settings/cloudmessaging/
	ServerKey string        `mapstructure:"key"`
	Retries   int           `mapstructure:"retries"`
	Timeout   time.Duration `mapstructure:"timeout"`
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

	if strings.TrimSpace(c.ServerKey) == "" {
		return nil, errors.New("invalid server key")
	}

	return c, nil
}
