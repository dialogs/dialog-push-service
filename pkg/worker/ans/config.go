package ans

import (
	"os"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`

	// Path to tls file in pem format
	PemFile     string        `mapstructure:"pem"`
	SendTries   int           `mapstructure:"send-tries"`
	SendTimeout time.Duration `mapstructure:"send-timeout"`
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

	if _, err := os.Stat(c.PemFile); err != nil {
		return nil, errors.Wrap(err, "ans: pem")
	}

	return c, nil
}
