package worker

import (
	"context"
	"runtime"

	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	ErrEmptyToken           = NewResponseError(ErrorCodeBadDeviceToken, errors.New("empty device token"))
	ErrUnknownResponseError = NewResponseError(ErrorCodeUnknown, errors.New("unknown response error"))
	ErrInvalidOutDataType   = NewResponseError(ErrorCodeBadRequest, errors.New("invalid out data type"))
)

type FnNewNotification func() interface{}
type FnSendNotification func(ctx context.Context, token string, out interface{}) error

type Worker struct {
	projectID          string
	kind               Kind
	nopMode            bool
	threads            chan struct{}
	logger             *zap.Logger
	metric             *metric.Provider
	reqConverter       converter.IRequestConverter
	fnNewNotification  FnNewNotification
	fnSendNotification FnSendNotification
}

func New(
	cfg *Config,
	kind Kind,
	logger *zap.Logger,
	svcMetric *metric.Service,
	reqConverter converter.IRequestConverter,
	fnNewNotification FnNewNotification,
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

	return &Worker{
		projectID:          cfg.ProjectID,
		kind:               kind,
		nopMode:            cfg.NopMode,
		threads:            threads,
		logger:             logger.With(zap.String("worker", kind.String())),
		metric:             providerMetric,
		reqConverter:       reqConverter,
		fnNewNotification:  fnNewNotification,
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

func (w *Worker) Send(ctx context.Context, req *Request) <-chan *Response {

	ch := make(chan *Response)
	reserved := <-w.threads

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

		out := w.fnNewNotification()
		err := w.reqConverter.Convert(req.Payload, out)

		for _, token := range req.Devices {
			resp := &Response{
				ProjectID:   w.projectID,
				DeviceToken: token,
			}

			// hide device token to hash
			l := w.logger.With(
				zap.String("token hash", TokenHash(token)),
				zap.String("id", req.CorrelationID))

			select {
			case <-ctx.Done():
				return
			default:
				if resp.DeviceToken == "" {
					l.Error("empty token")
					resp.Error = ErrEmptyToken

				} else if err != nil {
					// convert error
					l.Error("convert incoming message", zap.Error(err))
					resp.Error = err

				} else if w.nopMode {
					l.Info("nop mode", zap.Any("send notification", resp))

				} else {

					timerCancel := w.metric.NewIOTimer()
					err := w.fnSendNotification(ctx, token, out)
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
