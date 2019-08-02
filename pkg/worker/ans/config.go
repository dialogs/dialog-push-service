package ans

import (
	"os"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2ans"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	*worker.Config `mapstructure:"-"`
	APIConfig      *api2ans.Config `mapstructure:"-"`

	// Path to tls file in pem format
	PemFile string `mapstructure:"pem"`
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
		c.APIConfig, err = api2ans.NewConfig(src)
	case converter.KindBinary:
		// nothing do
	default:
		err = errors.New("invalid converter config kind")
	}

	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(c.PemFile); err != nil {
		return nil, errors.Wrap(err, "ans: pem")
	}

	return c, nil
}
