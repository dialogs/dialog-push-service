package ans

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/converter/api2ans"
	"github.com/dialogs/dialog-push-service/pkg/converter/binary"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

type Worker struct {
	*worker.Worker
	provider *ans.Client
}

func New(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (*Worker, error) {

	pem, err := ioutil.ReadFile(cfg.PemFile)
	if err != nil {
		return nil, err
	}

	provider, err := ans.NewFromPem(pem, cfg.IsSandbox)
	if err != nil {
		return nil, err
	}

	var reqConverter converter.IRequestConverter

	switch cfg.ConverterKind {
	case converter.KindApi:
		reqConverter, err = api2ans.NewRequestConverter(cfg.APIConfig, provider.Certificate())
		if err != nil {
			return nil, err
		}

	case converter.KindBinary:
		reqConverter = binary.NewRequestConverter()

	}

	w := &Worker{
		provider: provider,
	}

	w.Worker, err = worker.New(
		cfg.Config,
		worker.KindApns,
		provider.DevelopMode(),
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
	return &ans.Request{}
}

func (w *Worker) sendNotification(ctx context.Context, token string, out interface{}) error {

	req, ok := out.(*ans.Request)
	if !ok {
		return worker.ErrInvalidOutDataType
	}

	req.Token = url.QueryEscape(token)

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.StatusCode != 200 {
		err := errors.New(strconv.Itoa(answer.StatusCode) + " " + answer.Reason)
		if answer.StatusCode == 400 && answer.Reason == "BadDeviceToken" {
			return worker.NewResponseErrorBadDeviceToken(err)
		}

		return worker.NewResponseErrorFromAnswer(answer.StatusCode, err)
	}

	return nil
}
