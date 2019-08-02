package fcm

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2fcm"
	"github.com/dialogs/dialog-push-service/pkg/converter/binary"
	"github.com/dialogs/dialog-push-service/pkg/provider/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

type Worker struct {
	*worker.Worker
	provider *fcm.Client
}

func New(cfg *Config, logger *zap.Logger) (*Worker, error) {

	if cfg.SendTries <= 0 {
		cfg.SendTries = 2
	}

	if cfg.SendTimeout <= 0 {
		cfg.SendTimeout = time.Second
	}

	serviceAccount, err := ioutil.ReadFile(cfg.ServiceAccount)
	if err != nil {
		return nil, err
	}

	provider, err := fcm.New(serviceAccount, cfg.SendTries, cfg.SendTimeout)
	if err != nil {
		return nil, err
	}

	var reqConverter converter.IRequestConverter

	switch cfg.ConverterKind {
	case converter.KindApi:
		reqConverter = api2fcm.NewRequestConverter(cfg.APIConfig)

	case converter.KindBinary:
		reqConverter = binary.NewRequestConverter()

	}

	w := &Worker{
		provider: provider,
	}

	kind := worker.KindFcm
	w.Worker = worker.New(
		cfg.Config,
		kind,
		logger.With(zap.String("worker", kind.String())),
		reqConverter,
		w.newNotification,
		w.sendNotification,
	)

	return w, nil
}

func (w *Worker) newNotification() interface{} {
	return &fcm.Request{}
}

func (w *Worker) sendNotification(ctx context.Context, token string, out interface{}) error {

	req, ok := out.(*fcm.Request)
	if !ok {
		return worker.ErrInvalidOutDataType
	}

	req.Message.Token = token

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.Error != nil {
		return answer.Error

	}

	return nil
}
