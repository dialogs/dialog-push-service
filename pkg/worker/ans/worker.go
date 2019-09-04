package ans

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/dialogs/dialog-push-service/pkg/provider/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

var ErrInvalidRequestType = errors.New("invalid apns request type")

type Worker struct {
	*worker.Worker
	provider *ans.Client
}

func New(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (*Worker, error) {

	fPem, err := os.Open(cfg.PemFile)
	if err != nil {
		return nil, err
	}
	defer fPem.Close()

	// SAST: exception 'utils.ReadFile prone to resource exhaustion'
	pemSize, err := fPem.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	} else if pemSize > 1024*1024*10 {
		return nil, fmt.Errorf("invalid pem file size: %d", pemSize)
	}

	if _, err := fPem.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	pem := bytes.NewBuffer(make([]byte, 0, pemSize))
	if _, err := io.Copy(pem, fPem); err != nil {
		return nil, err
	}

	provider, err := ans.NewFromPem(pem.Bytes(), cfg.Sandbox, cfg.Retries, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	w := &Worker{
		provider: provider,
	}

	w.Worker, err = worker.New(
		cfg.Config,
		worker.KindApns,
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

func (w *Worker) SupportsVoIP() bool {
	return w.provider.SupportsVoIP()
}

func (w *Worker) sendNotification(ctx context.Context, in provider.IRequest) error {

	req, ok := in.(*ans.Request)
	if !ok || req == nil {
		return ErrInvalidRequestType
	}

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.StatusCode != 200 {
		msg := answer.Body.Reason
		if msg == "" {
			msg = http.StatusText(answer.StatusCode)
		}

		err := errors.New(strconv.Itoa(answer.StatusCode) + " " + msg)
		if answer.StatusCode == http.StatusBadRequest && answer.Body.Reason == "BadDeviceToken" {
			return worker.NewResponseErrorBadDeviceToken(err)
		}

		return worker.NewResponseErrorFromAnswer(answer.StatusCode, err)
	}

	return nil
}
