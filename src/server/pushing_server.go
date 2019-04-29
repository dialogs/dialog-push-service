package main

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
	"io"

	"golang.org/x/net/context"
)

type PushingServerImpl struct {
	metricsCollector *metricsCollector
	providers        map[string]DeliveryProvider
	readQueueSize    int
	writeQueueSize   int
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

func (p *PushingServerImpl) startStream(ctx context.Context, requests chan *Push, responses chan<- *PushResult) {
	responder := NewStreamResponder(ctx, responses)
	for push := range requests {
		log.Infof("Incoming streaming request: %s", push.GoString())
		p.deliverPush(push, responder)
	}
}

func (p *PushingServerImpl) Ping(ctx context.Context, ping *PingRequest) (*PongResponse, error) {
	return &PongResponse{}, nil
}

func streamOut(stream Pushing_PushStreamServer, responses <-chan *PushResult, errch chan<- error) {
	log.Infof("Opening stream out")
	defer func() { log.Infof("Closing stream out") }()

	for {
		select {
		case res := <-responses:
			response := &Response{
				ProjectInvalidations: map[string]*DeviceIdList{res.ProjectId: res.Failures},
			}
			err := stream.Send(response)
			if err != nil {
				errch <- err
				return
			}
		case <-stream.Context().Done():
			return
		}
	}
}

func streamIn(pm *peerMetrics, stream Pushing_PushStreamServer, requests chan<- *Push, errch chan<- error) {
	log.Infof("Opening stream in")
	defer func() { log.Infof("Closing stream in") }()

	for {
		request, err := stream.Recv()
		if err != nil {
			errch <- err
			break
		}
		if request == nil {
			log.Info("Empty push, skipping")
			continue
		}
		pm.pushRecv.Inc()
		log.Infof("Incoming push request: %s", request.GoString())
		requests <- request
	}

	close(requests)
}

func (p *PushingServerImpl) getAddrInfo(ctx context.Context) string {
	peer, peerOk := peer.FromContext(ctx)
	if peerOk {
		return peer.Addr.String()
	}

	return "unknown address"
}

func (p *PushingServerImpl) PushStream(stream Pushing_PushStreamServer) error {
	errch := make(chan error, 2)
	requests := make(chan *Push, p.readQueueSize)
	responses := make(chan *PushResult, p.writeQueueSize)

	addrInfo := p.getAddrInfo(stream.Context())
	pm, err := p.metricsCollector.getMetricsForPeer(addrInfo)
	if err != nil {
		return err
	}

	defer func() {
		// close(requests)
		// close(responses)
		// close(errch)
		log.Infof("Closing stream: %s", addrInfo)
	}()

	log.Infof("Starting stream: %s", addrInfo)
	go p.startStream(stream.Context(), requests, responses)
	go streamOut(stream, responses, errch)
	go streamIn(pm, stream, requests, errch)

	err = <-errch
	if err == nil || err == io.EOF {
		log.Infof("Stream completed normally: %s", addrInfo)
	} else {
		log.Errorf("Stopping stream %s due to error: %s", addrInfo, err.Error())
	}
	return err
}

type empty struct{}

func mergeDeviceLists(target *DeviceIdList, source *DeviceIdList) *DeviceIdList {
	set := make(map[string]empty)
	for _, n := range target.DeviceIds {
		set[n] = empty{}
	}
	for _, n := range source.DeviceIds {
		set[n] = empty{}
	}

	result := &DeviceIdList{DeviceIds: make([]string, 0, len(set))}
	for k, _ := range set {
		result.DeviceIds = append(result.DeviceIds, k)
	}
	return result
}

func mergeResponses(target, source *Response) {
	for k, v := range source.ProjectInvalidations {
		found, exists := target.ProjectInvalidations[k]
		if exists {
			target.ProjectInvalidations[k] = mergeDeviceLists(found, v)
		} else {
			target.ProjectInvalidations[k] = v
		}
	}
}

func (p *PushingServerImpl) SinglePush(ctx context.Context, push *Push) (*Response, error) {
	addrInfo := p.getAddrInfo(ctx)
	pm, err := p.metricsCollector.getMetricsForPeer(addrInfo)
	if err != nil {
		return nil, err
	}
	pm.pushRecv.Inc()

	rsp := &Response{ProjectInvalidations: make(map[string]*DeviceIdList)}
	response := make(chan *PushResult, p.writeQueueSize)
	responder := NewUnaryResponder(ctx, response)

	taskCount := p.deliverPush(push, responder)

	for i := 0; i < taskCount; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-response:
			rsp.ProjectInvalidations[res.ProjectId] = res.Failures

			if len(res.Failures.DeviceIds) > 0 {
				log.Infof("Invalidations for push with id = `%s` for projectId = `%s` is `%s`", push.CorrelationId, res.ProjectId, res.Failures)
			}
		}
	}

	return rsp, nil
}

func ensureProjectIdUniqueness(projectId string, providers map[string]DeliveryProvider) bool {
	_, exists := providers[projectId]
	return !exists
}

func newPushingServer(config *serverConfig) PushingServer {
	m := newMetricsCollector()
	p := &PushingServerImpl{
		metricsCollector: m,
		providers:        make(map[string]DeliveryProvider),
		readQueueSize:    config.ReadQueueSize,
		writeQueueSize:   config.WriteQueueSize,
	}
	for _, c := range config.getProviderConfigs() {
		if !ensureProjectIdUniqueness(c.getProjectID(), p.providers) {
			log.Fatalf("Duplicate project id: %s", c.getProjectID())
		}
		provider := c.newProvider()
		pm, err := m.getMetricsForProvider(c.getKind(), c.getProjectID())
		if err != nil {
			log.Fatalf("Failed to create metrics for provider %s: %s", c.getProjectID(), err.Error())
		}
		spawnWorkers(provider, pm)
		p.providers[c.getProjectID()] = provider
	}
	return p
}
