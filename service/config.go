package service

import (
	"errors"
	"fmt"

	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/dialogs/dialog-push-service/pkg/worker/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker/gcm"
	"github.com/spf13/viper"
)

type Config struct {
	Fcm       []*fcm.Config `mapstructure:"-"`
	Gcm       []*gcm.Config `mapstructure:"-"`
	Ans       []*ans.Config `mapstructure:"-"`
	ApiPort   string        `mapstructure:"grpc-port"`
	AdminPort string        `mapstructure:"http-port"`
}

func NewConfig(src *viper.Viper) (*Config, error) {

	c := &Config{}
	err := src.Unmarshal(c)
	if err != nil {
		return nil, err
	}

	c.Ans, err = getAppleConfig(src)
	if err != nil {
		return nil, err
	}

	c.Gcm, err = getGcmConfig(src)
	if err != nil {
		return nil, err
	}

	c.Fcm, err = getFcmConfig(src)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) WalkConfigs(fn func(interface{}) error) error {

	for _, item := range c.Ans {
		if err := fn(item); err != nil {
			return err
		}
	}

	for _, item := range c.Gcm {
		if err := fn(item); err != nil {
			return err
		}
	}

	for _, item := range c.Fcm {
		if err := fn(item); err != nil {
			return err
		}
	}

	return nil
}

func getFcmConfig(src *viper.Viper) ([]*fcm.Config, error) {

	srcList, err := getConfigListByKey(src, worker.KindFcm.String())
	if err != nil {
		return nil, err
	}

	retval := make([]*fcm.Config, 0, len(srcList))
	for _, item := range srcList {
		cfg, err := fcm.NewConfig(item)
		if err != nil {
			return nil, err
		}

		retval = append(retval, cfg)
	}

	return retval, nil
}

func getGcmConfig(src *viper.Viper) ([]*gcm.Config, error) {

	srcList, err := getConfigListByKey(src, worker.KindGcm.String())
	if err != nil {
		return nil, err
	}

	retval := make([]*gcm.Config, 0, len(srcList))
	for _, item := range srcList {
		cfg, err := gcm.NewConfig(item)
		if err != nil {
			return nil, err
		}

		retval = append(retval, cfg)
	}

	return retval, nil
}

func getAppleConfig(src *viper.Viper) ([]*ans.Config, error) {

	srcList, err := getConfigListByKey(src, worker.KindApns.String())
	if err != nil {
		return nil, err
	}

	retval := make([]*ans.Config, 0, len(srcList))
	for _, item := range srcList {
		cfg, err := ans.NewConfig(item)
		if err != nil {
			return nil, err
		}

		retval = append(retval, cfg)
	}

	return retval, nil
}

func getConfigListByKey(src *viper.Viper, key string) ([]*viper.Viper, error) {

	sub := src.Get(key)
	if sub == nil {
		return make([]*viper.Viper, 0), nil
	}

	arr, ok := sub.([]interface{})
	if !ok {
		return nil, errors.New("is not array:" + key)
	}

	retval := make([]*viper.Viper, 0, len(arr))
	for i, item := range arr {
		m, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid array item #%d: '%s'", i, key)
		}

		dest := viper.New()
		for k, v := range m {
			kStr, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("invalid key '%s.%v'", key, k)
			}

			dest.Set(kStr, v)
		}

		retval = append(retval, dest)
	}

	return retval, nil
}
