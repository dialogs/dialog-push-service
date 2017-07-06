package main

import (
	"github.com/spf13/viper"
	"os"
	"errors"
	"go.uber.org/zap"
)

type workersPool struct {
	Workers uint8
}

type sandboxing struct {
	IsSandbox bool `mapstructure:"sandbox"`
}

type apnsConfig struct {
	ProjectID   string `mapstructure:"project-id"`
	Topic       string
	PemFile     string `mapstructure:"pem"`
	IsVoip      bool   `mapstructure:"voip"`
	AllowAlerts bool   `mapstructure:"allow-alerts"`
	sandboxing  `mapstructure:",squash"`
	workersPool `mapstructure:",squash"`
}

type googleConfig struct {
	ProjectID   string `mapstructure:"project-id"`
	Key         string
	Retries     uint8
	AllowAlerts bool   `mapstructure:"allow-alerts"`
	sandboxing  `mapstructure:",squash"`
	workersPool `mapstructure:",squash"`
}

type providerConstructor interface {
	getProjectID() string
	newProvider(*zap.Logger) DeliveryProvider
}

func (g googleConfig) getProjectID() string {
	return g.ProjectID
}

func (a apnsConfig) getProjectID() string {
	return a.ProjectID
}

func (g googleConfig) checkConfig() (err error) {
	if len(g.ProjectID) == 0 {
		err = errors.New("No correct `project-id` found")
	}
	if len(g.Key) == 0 {
		err = errors.New("No correct `key` found")
	}
	return
}

func (a apnsConfig) checkConfig() (err error) {
	if len(a.ProjectID) == 0 {
		err = errors.New("No correct `project-id` found")
	}
	if len(a.PemFile) == 0 {
		err = errors.New("No correct `pem` found")
	}
	return
}

type serverConfig struct {
	Google   []googleConfig
	Apple    []apnsConfig
	Logger   *zap.Logger
	GrpcPort uint16 `mapstructure:"grpc-port"`
	HTTPPort uint16 `mapstructure:"http-port"`
}

func (c *serverConfig) getProviderConfigs() []providerConstructor {
	constructors := make([]providerConstructor, 0, len(c.Google)+len(c.Apple))
	for _, g := range c.Google {
		constructors = append(constructors, g)
	}
	for _, a := range c.Apple {
		constructors = append(constructors, a)
	}
	return constructors
}

func newConfig(logger *zap.Logger) *serverConfig {
	cfg := &serverConfig{Logger: logger}
	cfg.Google = make([]googleConfig, 0)
	cfg.Apple = make([]apnsConfig, 0)
	return cfg
}

func loadConfig(filename string, logger *zap.Logger) (config *serverConfig, err error) {
	var file *os.File
	viper.SetConfigType("YAML")
	if file, err = os.Open(filename); err != nil {
		return
	}
	if err = viper.ReadConfig(file); err != nil {
		return
	}
	config = newConfig(logger)
	if err = viper.Unmarshal(config); err != nil {
		return
	}
	for k := range config.Google {
		err = config.Google[k].checkConfig()
		if err != nil {
			return
		}
		if config.Google[k].Workers == 0 {
			config.Google[k].Workers = 1
		}
	}
	for k := range config.Apple {
		err = config.Apple[k].checkConfig()
		if err != nil {
			return
		}
		if config.Apple[k].Workers == 0 {
			config.Apple[k].Workers = 1
		}
	}
	return
}
