package main

import (
	"io"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/context"
)

type PushingServerImpl struct {
	providers map[string]DeliveryProvider
}

func peerTypeProtobufToMPS(peerType PeerType) int {
	switch peerType {
	case Private:
		return 1
	case Group:
		return 2
	case SIP:
		return 4
	default:
		return 0
	}
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
		log.Infof("Incoming request: %+v", req)
		for projectId, deviceList := range req.GetDestinations() {
			deviceIds := deviceList.GetDeviceIds()
			provider, exists := p.providers[projectId]
			if !exists {
				log.WithField("correlationId", req.CorrelationId).Errorf("No provider found for projectId: %s", projectId)
				continue
			}
			if len(deviceIds) == 0 {
				log.WithField("correlationId", req.CorrelationId).Infof("Empty deviceIds", req.CorrelationId)
				continue
			}
			if len(deviceIds) >= 1000 {
				log.WithField("correlationId", req.CorrelationId).Warnf(" DeviceIds array should contain at most 999 items", req.CorrelationId)
				continue
			}
			provider.getTasksChan() <- PushTask{deviceIds: deviceIds, body: req.GetBody(), resp: resps[projectId], correlationId: req.CorrelationId}
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

func streamIn(stream Pushing_PushStreamServer, requests chan *Push, errch chan error) {
	for {
		request, err := stream.Recv()
		if err != nil {
			errch <- err
			return
		}
		if request == nil {
			log.Info("Empty push, skipping")
			continue
		}
		requests <- request
	}
}

func (p PushingServerImpl) PushStream(stream Pushing_PushStreamServer) error {
	log.Info("Starting stream")
	errch := make(chan error)
	requests := make(chan *Push)
	responses := make(chan *Response)
	defer func() {
		close(requests)
		close(responses)
		close(errch)
		log.Infof("Closing stream")
	}()
	go p.startStream(requests, responses)
	go streamOut(stream, responses, errch)
	go streamIn(stream, requests, errch)
	err := <-errch
	if err == nil || err == io.EOF {
		log.Info("Stream completed normally")
	} else {
		log.Errorf("Stopping stream due to error: %s", err.Error())
	}
	return err
}

func ensureProjectIdUniqueness(projectId string, providers map[string]DeliveryProvider) bool {
	_, exists := providers[projectId]
	return !exists
}

func newPushingServer(config *serverConfig) PushingServer {
	p := PushingServerImpl{providers: make(map[string]DeliveryProvider)}
	for _, c := range config.getProviderConfigs() {
		if !ensureProjectIdUniqueness(c.getProjectID(), p.providers) {
			log.Fatalf("Duplicate project id: %s", c.getProjectID())
		}
		provider := c.newProvider()
		spawnWorkers(provider)
		p.providers[c.getProjectID()] = provider
	}
	return p
}
