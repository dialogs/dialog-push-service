package main

import (
	"io"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"

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

func workerOutputLoop(projectId string, rsp chan *Response, in chan *DeviceIdList) {
	log.Info("Opening output loop")
	for res := range in {
		// We don't need to send empty pseudo-acks in a stream mode
		if len(res.DeviceIds) > 0 {
			inv := make(map[string]*DeviceIdList, 1)
			inv[projectId] = res
			rsp <- &Response{ProjectInvalidations: inv}
		}
	}
	log.Info("Closing output loop")
}

func (p PushingServerImpl) getProvidersResponders() map[string]chan *DeviceIdList {
	resps := make(map[string]chan *DeviceIdList, len(p.providers))
	for projectId := range p.providers {
		resps[projectId] = make(chan *DeviceIdList, 1)
	}
	return resps
}

func (p PushingServerImpl) startStream(requests chan *Push, responses chan *Response) {
	resps := p.getProvidersResponders()
	for projectId, out := range resps {
		// TODO: make timed output with aggregated results? [groupedWithin]
		go workerOutputLoop(projectId, responses, out)
	}
	for push := range requests {
		log.Infof("Incoming streaming request: %s", push.GoString())
		p.deliverPush(push, resps)
	}
}

func (p PushingServerImpl) Ping(ctx context.Context, ping *PingRequest) (*PongResponse, error) {
	return &PongResponse{}, nil
}

func streamOut(stream Pushing_PushStreamServer, responses chan *Response, errch chan error) {
	log.Infof("Opening stream out")
	defer func() { log.Infof("Closing stream out") }()

	for resp := range responses {
		err := stream.Send(resp)
		if err != nil {
			errch <- err
			return
		}
	}
}

func streamIn(stream Pushing_PushStreamServer, requests chan *Push, errch chan error) {
	log.Infof("Opening stream in")
	defer func() { log.Infof("Closing stream in") }()

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
	errch := make(chan error)
	requests := make(chan *Push)
	responses := make(chan *Response)

	var addrInfo string
	peer, peerOk := peer.FromContext(stream.Context())
	if peerOk {
		addrInfo = peer.Addr.String()
	} else {
		addrInfo = "unknown address"
	}
	defer func() {
		// close(requests)
		// close(responses)
		// close(errch)
		log.Infof("Closing stream: %s", addrInfo)
	}()

	log.Infof("Starting stream: %s", addrInfo)
	go p.startStream(requests, responses)
	go streamOut(stream, responses, errch)
	go streamIn(stream, requests, errch)

	err := <-errch
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

func (p PushingServerImpl) SinglePush(ctx context.Context, push *Push) (*Response, error) {
	responders := p.getProvidersResponders()
	var wg sync.WaitGroup
	rsp := &Response{ProjectInvalidations: make(map[string]*DeviceIdList)}

	mlock := sync.Mutex{}
	wg.Add(p.deliverPush(push, responders))
	for projectId, ch := range responders {
		go func(pid string, taskChan chan *DeviceIdList) {
			//log.Infof("Push: %d, chan: %+v", push.Body.Seq, taskChan)
			invalidations := <-taskChan
			if invalidations != nil {
				mlock.Lock()
				rsp.ProjectInvalidations[pid] = invalidations
				mlock.Unlock()
				if len(invalidations.DeviceIds) > 0 {
					log.Infof("Invalidations for push with id = `%s` for projectId = `%s` is `%s`", push.CorrelationId, pid, invalidations)
				}
				wg.Done()
			}
		}(projectId, ch)
	}

	wg.Wait()
	return rsp, nil
}

func ensureProjectIdUniqueness(projectId string, providers map[string]DeliveryProvider) bool {
	_, exists := providers[projectId]
	return !exists
}

func newPushingServer(config *serverConfig) PushingServer {
	m := newMetricsCollector()
	p := PushingServerImpl{providers: make(map[string]DeliveryProvider)}
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
