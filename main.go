package main

import (
	"log"
	"os"

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

	logCfg := zap.NewProductionConfig()
	logger, err := logCfg.Build()
	if err != nil {
		log.Fatal("failed to create logger:", err)
	}

	v := viper.New()
	v.SetConfigFile(opts.ConfigLocation)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("failed to parse config:", err)
	}

	svc, err := service.New(v, logger)
	if err != nil {
		log.Fatal("failed to parse config:", err)
	}

	if err := svc.Run(); err != nil {
		log.Println("close service", err)
	}
}
