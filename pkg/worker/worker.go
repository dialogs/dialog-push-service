package worker

import (
	"context"
	"runtime"

	"github.com/dialogs/dialog-push-service/pkg/conversion"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	ErrEmptyToken           = NewResponseError(ErrorCodeBadDeviceToken, errors.New("empty device token"))
	ErrUnknownResponseError = NewResponseError(ErrorCodeUnknown, errors.New("unknown response error"))
	ErrInvalidOutDataType   = NewResponseError(ErrorCodeBadRequest, errors.New("invalid out data type"))
)

type FnSendNotification func(ctx context.Context, out provider.IRequest) error

type Worker struct {
	projectID          string
	kind               Kind
	nopMode            bool
	threads            chan struct{}
	logger             *zap.Logger
	metric             *metric.Provider
	conversionConfig   conversion.Config
	fnSendNotification FnSendNotification
}

func New(
	cfg *Config,
	kind Kind,
	sandbox bool,
	logger *zap.Logger,
	svcMetric *metric.Service,
	fnSendNotification FnSendNotification,
) (*Worker, error) {

	countThreads := cfg.CountThreads
	if countThreads <= 0 {
		countThreads = runtime.NumCPU()
	}

	threads := make(chan struct{}, countThreads)
	for i := 0; i < countThreads; i++ {
		threads <- struct{}{}
	}

	providerMetric, err := svcMetric.GetProviderMetrics(kind.String(), cfg.ProjectID)
	if err != nil {
		return nil, err
	}

	l := logger.With(
		zap.String("worker", kind.String()),
		zap.String("project ID", cfg.ProjectID))
	if sandbox {
		l = l.With(zap.Bool("develop", sandbox))
	}

	return &Worker{
		projectID:          cfg.ProjectID,
		kind:               kind,
		nopMode:            cfg.NopMode,
		threads:            threads,
		logger:             l,
		conversionConfig:   *cfg.Config,
		metric:             providerMetric,
		fnSendNotification: fnSendNotification,
	}, nil
}

func (w *Worker) Kind() Kind {
	return w.kind
}

func (w *Worker) ProjectID() string {
	return w.projectID
}

func (w *Worker) NoOpMode() bool {
	return w.nopMode
}

func (w *Worker) ConversionConfig() *conversion.Config {
	return &w.conversionConfig
}

func (w *Worker) Send(ctx context.Context, req *Request) <-chan *Response {

	ch := make(chan *Response)
	reserved := <-w.threads
	// TODO: add wait timeout. if timeout is end, write to storage for retry

	go func() {
		defer func() { w.threads <- reserved }()
		defer close(ch)

		if len(req.Devices) == 0 {
			w.logger.Error(ErrEmptyToken.Error())

			ch <- &Response{
				ProjectID: w.projectID,
				Error:     ErrEmptyToken,
			}
			return
		}

		for _, token := range req.Devices {
			resp := &Response{
				ProjectID:   w.projectID,
				DeviceToken: token,
			}

			// hide device token
			tokenInfo := ""
			tokenPartLen := len(token) / 3
			if tokenPartLen > 0 {
				tokenInfo = token[:tokenPartLen] + "..." + token[len(token)-tokenPartLen:]
			}

			l := w.logger.With(
				zap.String("token", tokenInfo),
				zap.String("id", req.CorrelationID))

			select {
			case <-ctx.Done():
				return
			default:
				if resp.DeviceToken == "" {
					l.Error("empty token")
					resp.Error = ErrEmptyToken

				} else if w.nopMode {
					l.Info("nop mode", zap.Any("send notification", resp))

				} else {
					req.Payload.SetToken(token)

					timerCancel := w.metric.NewIOTimer()
					err := w.fnSendNotification(ctx, req.Payload)
					timerCancel()

					if err != nil {
						w.metric.FailsInc()
						resp.Error = err
						l.Error("failed to send", zap.Error(resp.Error))
					} else {
						w.metric.SuccessInc()
						l.Info("success send")
					}
				}
			}

			ch <- resp
		}
	}()

	return ch
}
