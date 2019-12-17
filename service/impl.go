package service

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/conversion"
	"github.com/dialogs/dialog-push-service/pkg/metric"
	"github.com/dialogs/dialog-push-service/pkg/worker"
	"github.com/dialogs/dialog-push-service/pkg/worker/ans"
	"github.com/dialogs/dialog-push-service/pkg/worker/fcm"
	"github.com/dialogs/dialog-push-service/pkg/worker/gcm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/peer"
)

var (
	errInvalidProjectID         = errors.New("invalid project ID")
	errUnknownConversationRules = errors.New("unknown conversation rules pb to provider request")
)

type implGRPC struct {
	metric  *metric.Service
	workers map[string]worker.IWorker
	logger  *zap.Logger
}

func newImplGRPC(cfg *Config, logger *zap.Logger) (*implGRPC, error) {

	svcMetric := metric.New()

	workers, err := getWorkers(cfg, logger, svcMetric)
	if err != nil {
		return nil, err
	}

	return &implGRPC{
		metric:  svcMetric,
		workers: workers,
		logger:  logger,
	}, nil
}

// Remove seq from push if it contains encrypted body [DP-3327]
func cleanPush(push *api.Push) {
	if body := push.GetBody(); body != nil && body.GetEncryptedPush() != nil {
		body.Seq = 0
	}
}

func (i *implGRPC) Ping(context.Context, *api.PingRequest) (*api.PongResponse, error) {
	return &api.PongResponse{}, nil
}

func (i *implGRPC) PushStream(stream api.Pushing_PushStreamServer) error {

	l := i.logger.With(zap.String("method", "push stream"))
	defer func() { l.Info("close stream") }()

	for {
		push, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		go func(taskLogger *zap.Logger, task *api.Push) {
			chOut, err := i.sendPush(stream.Context(), task, l)
			if err != nil {
				taskLogger.Error("failed to send push", zap.Error(err))
				return
			}

			for pushRes := range chOut {
				if len(pushRes.InvalidationDevices) == 0 {
					taskLogger.Info("empty invalidation devices list", zap.String("project id", pushRes.ProjectID))
					continue
				}

				res := &api.Response{
					ProjectInvalidations: map[string]*api.DeviceIdList{
						pushRes.ProjectID: &api.DeviceIdList{
							DeviceIds: pushRes.InvalidationDevices,
						},
					},
				}

				taskLogger.Info("send: start")
				if err := stream.Send(res); err != nil {
					l.Error("send: error", zap.Error(err))
				} else {
					taskLogger.Info("send: end")
				}
			}

		}(l.With(zap.String("correlation id", push.CorrelationId)), push)
	}
}

func (i *implGRPC) SinglePush(ctx context.Context, push *api.Push) (*api.Response, error) {

	l := i.logger.With(zap.String("method", "single push"))

	cleanPush(push)

	chRes, err := i.sendPush(ctx, push, l)
	if err != nil {
		return nil, err
	}

	res := &api.Response{
		ProjectInvalidations: make(map[string]*api.DeviceIdList, len(push.Destinations)),
	}

	for pushRes := range chRes {
		target, ok := res.ProjectInvalidations[pushRes.ProjectID]
		if ok {
			target.DeviceIds = append(target.DeviceIds, pushRes.InvalidationDevices...)
		} else {
			target = &api.DeviceIdList{
				DeviceIds: pushRes.InvalidationDevices,
			}
		}

		res.ProjectInvalidations[pushRes.ProjectID] = target
	}

	return res, nil
}

func (i *implGRPC) sendPush(ctx context.Context, push *api.Push, l *zap.Logger) (<-chan *sendPushResult, error) {

	l = l.With(zap.String("id", push.CorrelationId))

	addrInfo := i.getAddrInfo(ctx)
	peerMetric, err := i.metric.GetPeerMetrics(addrInfo)
	if err != nil {
		l.Error("get peer metric", zap.Error(err))
		return nil, err
	}

	cleanPush(push)

	peerMetric.Inc()

	chOut := make(chan *sendPushResult)

	go func() {
		defer func() { close(chOut) }()

		if len(push.Destinations) == 0 {
			return
		}

		wg := sync.WaitGroup{}

		for projectID, deviceList := range push.Destinations {
			projectLogger := l.With(zap.String("project id", projectID))

			w, err := i.getWorker(projectID)
			if err != nil {
				projectLogger.Error("get worker", zap.Error(err))

				chOut <- newSendPushResult(projectID)
				continue
			}

			wg.Add(1)
			go func(projectWorker worker.IWorker, devices []string) {
				defer wg.Done()

				req := &worker.Request{
					Devices:       devices,
					CorrelationID: push.CorrelationId,
				}

				conversationConfig := projectWorker.ConversionConfig()

				switch w.Kind() {
				case worker.KindApns:
					req.Payload, err = conversion.RequestPbToAns(push.Body, w.SupportsVoIP(), conversationConfig.AllowAlerts, &conversationConfig.Topic, &conversationConfig.Sound)
				case worker.KindFcm:
					req.Payload, err = conversion.RequestPbToFcm(push.Body, conversationConfig.AllowAlerts)
				case worker.KindGcm:
					req.Payload, err = conversion.RequestPbToGcm(push.Body, conversationConfig.AllowAlerts)
				default:
					err = errUnknownConversationRules
				}

				if err != nil {
					projectLogger.Error("conversation", zap.Error(err))
				}

				pushRes := newSendPushResult(projectWorker.ProjectID())

				if err == nil && !req.Payload.ShouldIgnore() {
					for res := range projectWorker.Send(ctx, req) {

						if res.Error != nil {
							workerErr, ok := res.Error.(*worker.ResponseError)

							if ok && (workerErr.Code == worker.ErrorCodeBadDeviceToken) {
								pushRes.InvalidationDevices = append(pushRes.InvalidationDevices, res.DeviceToken)
							}
						}
					}
				}

				chOut <- pushRes

			}(w, deviceList.GetDeviceIds())
		}

		wg.Wait()
	}()

	return chOut, nil
}

func (i *implGRPC) getWorker(projectID string) (worker.IWorker, error) {

	w, ok := i.workers[projectID]
	if !ok {
		return nil, errInvalidProjectID
	}

	return w, nil
}

func (i *implGRPC) getAddrInfo(ctx context.Context) string {
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
			wConf := c.(*ans.Config)
			w, err = ans.New(wConf, logger, svcMetric)
			if err != nil {
				err = errors.Wrap(err, "project ID: "+wConf.ProjectID)
			}

		case *gcm.Config:
			wConf := c.(*gcm.Config)
			w, err = gcm.New(wConf, logger, svcMetric)
			if err != nil {
				err = errors.Wrap(err, "project ID: "+wConf.ProjectID)
			}

		case *fcm.Config:
			wConf := c.(*fcm.Config)
			w, err = fcm.New(wConf, logger, svcMetric)
			if err != nil {
				err = errors.Wrap(err, "project ID: "+wConf.ProjectID)
			}

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
