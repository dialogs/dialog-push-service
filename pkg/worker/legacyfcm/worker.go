package legacyfcm

import (
	"context"
	"errors"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2legacyfcm"
	"github.com/dialogs/dialog-push-service/pkg/converter/binary"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider/legacyfcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

type Worker struct {
	*worker.Worker
	provider *legacyfcm.Client
}

func New(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (*Worker, error) {

	if cfg.SendTries <= 0 {
		cfg.SendTries = 2
	}

	provider, err := legacyfcm.New(cfg.ServerKey, cfg.SendTries, cfg.SendTimeout)
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

	w.Worker, err = worker.New(
		cfg.Config,
		worker.KindFcmLegacy,
		false,
		logger,
		svcMetric,
		reqConverter,
		w.newNotification,
		w.sendNotification,
	)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Worker) newNotification() interface{} {
	return &legacyfcm.Request{}
}

func (w *Worker) sendNotification(ctx context.Context, token string, out interface{}) error {

	req, ok := out.(*legacyfcm.Request)
	if !ok {
		return worker.ErrInvalidOutDataType
	}

	req.To = token

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.Success == 0 {
		var answerError error
		if len(answer.Results) > 0 {
			var errCode string
			for _, res := range answer.Results {
				errCode = res.Error
			}

			if errCode == legacyfcm.ErrorCodeInvalidRegistration ||
				errCode == legacyfcm.ErrorCodeMissingRegistration {

				return worker.NewResponseErrorBadDeviceToken(errors.New(errCode))
			}

			answerError = errors.New(strconv.Itoa(answer.StatusCode) + " " + errCode)

		} else {
			answerError = worker.ErrUnknownResponseError
		}

		return answerError

	}

	return nil
}
