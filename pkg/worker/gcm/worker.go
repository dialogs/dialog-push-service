package gcm

import (
	"context"
	"errors"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/dialogs/dialog-push-service/pkg/provider/gcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

var ErrInvalidRequestType = errors.New("invalid gcm request type")

type Worker struct {
	*worker.Worker
	provider *gcm.Client
}

func New(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (*Worker, error) {

	if cfg.SendTries <= 0 {
		cfg.SendTries = 2
	}

	provider, err := gcm.New([]byte(cfg.ServerKey), cfg.Sandbox, cfg.SendTries, cfg.SendTimeout)
	if err != nil {
		return nil, err
	}

	w := &Worker{
		provider: provider,
	}

	w.Worker, err = worker.New(
		cfg.Config,
		worker.KindGcm,
		provider.Sandbox(),
		logger,
		svcMetric,
		w.sendNotification,
	)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Worker) ExistVoIP() bool {
	return false
}

func (w *Worker) sendNotification(ctx context.Context, in provider.IRequest) error {

	req, ok := in.(*gcm.Request)
	if !ok {
		return ErrInvalidRequestType
	}

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.Success == 0 {
		var answerError error
		if len(answer.Results) > 0 {
			var errCode string
			for _, res := range answer.Results {
				errCode = res.Error
				break
			}

			if errCode == gcm.ErrorCodeInvalidRegistration ||
				errCode == gcm.ErrorCodeMissingRegistration {

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
