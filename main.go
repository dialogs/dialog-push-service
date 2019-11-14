package main

import (
	"log"
	"os"

	"github.com/dialogs/dialog-go-lib/logger"
	"github.com/dialogs/dialog-push-service/service"
	"github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var opts struct {
	ConfigLocation string `short:"c" long:"config" description:"Config file location" required:"true"`
}

func main() {

	if _, err := flags.ParseArgs(&opts, os.Args); err != nil {
		log.Fatal("failed to parse arguments:", err)
	}

	l, err := logger.New()
	if err != nil {
		log.Fatal("failed to create logger:", err)
	}

	l.Info("run service", zap.Any("info", service.Info()))

	v := viper.New()
	v.SetConfigFile(opts.ConfigLocation)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("failed to parse config:", err)
	}

	svc, err := service.New(v, l)
	if err != nil {
		log.Fatal("failed to parse config:", err)
	}

	if err := svc.Run(); err != nil {
		log.Println("close service", err)
	}
}
