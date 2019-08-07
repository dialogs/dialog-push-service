package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/dialogs/dialog-push-service/pkg/worker/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker/legacyfcm"
	"go.uber.org/zap"
	"google.golang.org/grpc/peer"
)

var errInvalidProjectID = errors.New("invalid project ID")

type impl struct {
	metric  *metric.Service
	workers map[string]worker.IWorker
	logger  *zap.Logger
}

func newImpl(cfg *Config, logger *zap.Logger) (*impl, error) {

	svcMetric := metric.New()

	workers, err := getWorkers(cfg, logger, svcMetric)
	if err != nil {
		return nil, err
	}

	return &impl{
		metric:  svcMetric,
		workers: workers,
		logger:  logger,
	}, nil
}

func (i *impl) Ping(context.Context, *api.PingRequest) (*api.PongResponse, error) {
	return &api.PongResponse{}, nil
}

func (i *impl) PushStream(stream api.Pushing_PushStreamServer) error {

	l := i.logger.With(zap.String("method", "push stream"))
	defer func() { l.Info("close stream") }()

	for {
		push, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		resp, err := i.sendPush(stream.Context(), push, l)
		if err != nil {
			return err
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (i *impl) SinglePush(ctx context.Context, push *api.Push) (*api.Response, error) {

	l := i.logger.With(zap.String("method", "single push"))

	return i.sendPush(ctx, push, l)
}

func (i *impl) sendPush(ctx context.Context, push *api.Push, l *zap.Logger) (*api.Response, error) {

	l = l.With(zap.String("id", push.CorrelationId))

	addrInfo := i.getAddrInfo(ctx)
	peerMetric, err := i.metric.GetPeerMetrics(addrInfo)
	if err != nil {
		l.Error("get peer metric", zap.Error(err))
		return nil, err
	}

	peerMetric.Inc()

	var (
		retval = &api.Response{
			ProjectInvalidations: make(map[string]*api.DeviceIdList, len(push.Destinations)),
		}
		retvalMu = sync.Mutex{}
		wg       = sync.WaitGroup{}
	)

	if len(push.Destinations) > 0 {

		for projectID, deviceList := range push.Destinations {

			w, err := i.getWorker(projectID)
			if err != nil {
				retvalMu.Lock()
				retval.ProjectInvalidations[projectID] = &api.DeviceIdList{}
				retvalMu.Unlock()

				l.Error("get worker", zap.Error(err), zap.String("project-id", projectID))
				continue
			}

			wg.Add(1)
			go func(projectWorker worker.IWorker, devices []string) {
				defer wg.Done()

				req := &worker.Request{
					Devices:       devices,
					CorrelationID: push.CorrelationId,
					Payload:       push.Body,
				}

				invalidations := &api.DeviceIdList{}

				for res := range projectWorker.Send(ctx, req) {

					if res.Error != nil {
						workerErr, ok := res.Error.(*worker.ResponseError)

						if ok && (workerErr.Code == worker.ErrorCodeBadDeviceToken) {
							invalidations.DeviceIds = append(invalidations.DeviceIds, res.DeviceToken)
						}
					}
				}

				retvalMu.Lock()
				retval.ProjectInvalidations[projectWorker.ProjectID()] = invalidations
				retvalMu.Unlock()

			}(w, deviceList.GetDeviceIds())
		}
	}

	wg.Wait()

	return retval, nil
}

func (i *impl) getWorker(projectID string) (worker.IWorker, error) {

	w, ok := i.workers[projectID]
	if !ok {
		return nil, errInvalidProjectID
	}

	return w, nil
}

func (i *impl) getAddrInfo(ctx context.Context) string {
	peer, peerOk := peer.FromContext(ctx)
	if peerOk {
		return peer.Addr.String()
	}

	return "unknown address"
}

func getWorkers(cfg *Config, logger *zap.Logger, svcMetric *metric.Service) (map[string]worker.IWorker, error) {

	m := make(map[string]worker.IWorker)

	err := cfg.WalkConfigs(func(c interface{}) error {
		var (
			w   worker.IWorker
			err error
		)

		switch c.(type) {
		case *ans.Config:
			w, err = ans.New(c.(*ans.Config), logger, svcMetric)
		case *legacyfcm.Config:
			w, err = legacyfcm.New(c.(*legacyfcm.Config), logger, svcMetric)
		case *fcm.Config:
			w, err = fcm.New(c.(*fcm.Config), logger, svcMetric)
		default:
			err = fmt.Errorf("unknown config type: %T", c)
		}

		if err != nil {
			return err
		}

		projectID := w.ProjectID()
		_, ok := m[projectID]
		if ok {
			return errors.New("not unique project id of a worker:" + projectID)
		}

		m[projectID] = w
		return nil
	})

	if err != nil {
		return nil, err
	}

	return m, nil
}
