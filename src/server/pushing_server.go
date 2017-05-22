package server

import (
	"golang.org/x/net/context"
	"io"
	"go.uber.org/zap"
)

type PushingServerImpl struct {
	providers map[string]DeliveryProvider
	logger *zap.Logger
}

func workerOutputLoop(projectId string, rsp chan *Response, in chan []string) {
	for res := range in {
		inv := make(map[string]*DeviceIdList, 1)
		inv[projectId] = &DeviceIdList{DeviceIds: res}
		rsp <- &Response{ProjectInvalidations: inv}
	}
}

func (p PushingServerImpl) startStream(requests chan *Push, responses chan *Response) {
	resps := make(map[string]chan []string, len(p.providers))
	defer func() {
		for _, ch := range resps {
			close(ch)
		}
	}()
	for projectId := range p.providers {
		out := make(chan []string)
		resps[projectId] = out
		// TODO: make timed output with aggregated results? [groupedWithin]
		go workerOutputLoop(projectId, responses, out)
	}
	for req := range requests {
		for projectId, deviceList := range req.GetDestinations() {
			deviceIds := deviceList.GetDeviceIds()
			if len(deviceIds) == 0 {
				p.logger.Info("Empty deviceIds")
				continue
			}
			if len(deviceIds) >= 1000 {
				p.logger.Warn("DeviceIds array should contain at most 999 items")
				continue
			}
			if provider, exists := p.providers[projectId]; !exists {
				p.logger.Error("No provider found for projectId", zap.String("projectId", projectId))
			} else {
				provider.getTasksChan() <- PushTask{deviceIds: deviceIds, body: req.GetBody(), resp: resps[projectId]}
			}
		}
	}
}

func (p PushingServerImpl) Ping(ctx context.Context, ping *PingRequest) (*PongResponse, error) {
	return &PongResponse{}, nil
}

func streamOut(stream Pushing_PushStreamServer, responses chan *Response, errch chan error) {
	for resp := range responses {
		err := stream.Send(resp)
		if err != nil {
			errch <- err
			return
		}
	}
}

func streamIn(stream Pushing_PushStreamServer, requests chan *Push, errch chan error, logger *zap.Logger) {
	for {
		request, err := stream.Recv()
		if err != nil {
			errch <- err
			return
		}
		if request == nil {
			logger.Info("Empty push, skipping")
			continue
		}
		requests <- request
	}
}

func (p PushingServerImpl) PushStream(stream Pushing_PushStreamServer) error {
	p.logger.Info("Starting stream")
	errch := make(chan error)
	requests := make(chan *Push)
	responses := make(chan *Response)
	defer func() {
		close(requests)
		close(responses)
		close(errch)
		p.logger.Info("Closing stream")
	}()
	go p.startStream(requests, responses)
	go streamOut(stream, responses, errch)
	go streamIn(stream, requests, errch, p.logger)
	err := <- errch
	if err == nil || err == io.EOF {
		p.logger.Info("Stream completed normally")
	} else {
		p.logger.Error("Stopping stream due to error", zap.Error(err))
	}
	return err
}

func ensureProjectIdUniqueness(projectId string, providers map[string]DeliveryProvider) bool {
	if _, exists := providers[projectId]; exists {
		return false
	}
	return true
}

func newPushingServer(config *serverConfig) PushingServer {
	p := PushingServerImpl{providers: make(map[string]DeliveryProvider), logger: config.Logger}
	for _, c := range config.getProviderConfigs() {
		if !ensureProjectIdUniqueness(c.getProjectID(), p.providers) {
			config.Logger.Fatal("Duplicate project id", zap.String("projectId", c.getProjectID()))
		}
		provider := c.newProvider(config.Logger)
		spawnWorkers(provider)
		p.providers[c.getProjectID()] = provider
	}
	return p
}
