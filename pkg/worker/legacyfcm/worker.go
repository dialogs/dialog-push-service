package legacyfcm

import (
	"context"
	"errors"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2legacyfcm"
	"github.com/dialogs/dialog-push-service/pkg/converter/binary"
	"github.com/dialogs/dialog-push-service/pkg/provider/legacyfcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/edganiukov/fcm"
	"go.uber.org/zap"
)

var ErrUnknownResponseError = errors.New("unknown response error")

type Worker struct {
	*worker.Worker
	provider *legacyfcm.Client
}

func New(cfg *Config, logger *zap.Logger) (*Worker, error) {

	if cfg.SendTries <= 0 {
		cfg.SendTries = 2
	}

	provider, err := legacyfcm.New(cfg.ServerKey, cfg.SendTries)
	if err != nil {
		return nil, err
	}

	var reqConverter converter.IRequestConverter

	switch cfg.ConverterKind {
	case converter.KindApi:
		reqConverter = api2legacyfcm.NewRequestConverter(cfg.APIConfig)

	case converter.KindBinary:
		reqConverter = binary.NewRequestConverter()

	}

	w := &Worker{
		provider: provider,
	}

	kind := worker.KindFcmLegacy
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
	return &fcm.Message{}
}

func (w *Worker) sendNotification(ctx context.Context, token string, out interface{}) error {

	req, ok := out.(*fcm.Message)
	if !ok {
		return worker.ErrInvalidOutDataType
	}

	req.Token = token

	answer, err := w.provider.Send(req)
	if err != nil {
		return err

	} else if answer.Success == 0 {
		var answerError error
		if len(answer.Results) > 0 {
			answerError = answer.Results[0].Error
		} else {
			answerError = ErrUnknownResponseError
		}

		return answerError

	}

	return nil
}
