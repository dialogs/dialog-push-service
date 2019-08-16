package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/dialogs/dialog-push-service/pkg/provider/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"go.uber.org/zap"
)

var ErrInvalidRequestType = errors.New("invalid fcm request type")

type Worker struct {
	*worker.Worker
	provider *fcm.Client
}

func New(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (*Worker, error) {

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

	provider, err := fcm.New(serviceAccount, cfg.Sandbox, cfg.SendTries, cfg.SendTimeout)
	if err != nil {
		return nil, err
	}

	w := &Worker{
		provider: provider,
	}

	w.Worker, err = worker.New(
		cfg.Config,
		worker.KindFcm,
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

	req, ok := in.(*fcm.Message)
	if !ok || req == nil {
		return ErrInvalidRequestType
	}

	answer, err := w.provider.Send(ctx, req)
	if err != nil {
		return err

	} else if answer.Error != nil {
		if answer.Error.Code == 400 && answer.Error.Status == fcm.ErrorCodeInvalidArgument {

			fields := getStringValueFromJSON(answer.Error.Details, "field")
			for i := range fields {
				if fields[i] == "message.token" {
					return worker.NewResponseErrorBadDeviceToken(answer.Error)
				}
			}
		}

		return answer.Error

	}

	return nil
}

func getStringValueFromJSON(src json.RawMessage, key string) []string {

	type State int
	const (
		StateReadObject State = iota
		StateReadKey
		StateReadValue
	)

	state := StateReadObject
	retval := make([]string, 0)
	if len(src) == 0 {
		return retval
	}

	r := bytes.NewReader(src)
	dec := json.NewDecoder(r)

	for {
		token, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil
		}

		switch token.(type) {
		case json.Delim:
			delim := token.(json.Delim)
			if delim == '{' {
				state = StateReadKey
			} else {
				state = StateReadObject
			}

		case string:
			val := token.(string)

			if state == StateReadKey && val == key {
				state = StateReadValue

			} else if state == StateReadValue {
				retval = append(retval, val)
				state = StateReadKey

			} else {
				state = StateReadKey

			}

		default:
			if state == StateReadValue {
				state = StateReadKey
			}
		}
	}

	return retval
}
