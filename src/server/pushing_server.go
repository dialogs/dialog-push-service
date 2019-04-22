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

func workerOutput(ctx context.Context, projectId string, rsp chan<- *Response, in <-chan *DeviceIdList) {
	select {
	case res := <-in:
		inv := make(map[string]*DeviceIdList, 1)
		inv[projectId] = res
		select {
		case rsp <- &Response{ProjectInvalidations: inv}:
		case <-ctx.Done():
			break
		}
	case <-ctx.Done():
	}
}

func (p PushingServerImpl) getProvidersResponders(push *Push) map[string]chan *DeviceIdList {
	projectIds := make([]string, 0, len(push.GetDestinations()))
	for projectId := range push.GetDestinations() {
		projectIds = append(projectIds, projectId)
	}
	resps := make(map[string]chan *DeviceIdList, len(p.providers))
	for _, projectId := range projectIds {
		if _, ok := p.providers[projectId]; ok {
			resps[projectId] = make(chan *DeviceIdList, 1)
		}
	}
	return resps
}

func (p PushingServerImpl) startStream(ctx context.Context, requests <-chan *Push, responses chan<- *Response) {
	for push := range requests {
		resps := p.getProvidersResponders(push)
		for projectId, out := range resps {
			go workerOutput(ctx, projectId, responses, out)
		}
		log.Infof("Incoming streaming request: %s", push.GoString())
		p.deliverPush(push, resps)
	}
}

func (p PushingServerImpl) Ping(ctx context.Context, ping *PingRequest) (*PongResponse, error) {
	return &PongResponse{}, nil
}

func streamOut(stream Pushing_PushStreamServer, responses <-chan *Response, errch chan<- error) {
	log.Infof("Opening stream out")
	defer func() { log.Infof("Closing stream out") }()

	for {
		select {
		case <-stream.Context().Done():
			errch <- stream.Context().Err()
		case resp := <-responses:
			err := stream.Send(resp)
			if err != nil {
				errch <- err
				return
			}
		}
	}

}

func streamIn(stream Pushing_PushStreamServer, requests chan<- *Push, errch chan<- error) {
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
		requests <- request
	}

	close(requests)
}

func (p PushingServerImpl) PushStream(stream Pushing_PushStreamServer) error {
	errch := make(chan error, 2)
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
	go p.startStream(stream.Context(), requests, responses)
	go streamOut(stream, responses, errch)
	go streamIn(stream, requests, errch)
	select {
	case <-stream.Context().Done():
		return nil
	case err := <-errch:
		if err == nil || err == io.EOF {
			log.Infof("Stream completed normally: %s", addrInfo)
		} else {
			log.Errorf("Stopping stream %s due to error: %s", addrInfo, err.Error())
		}
		return err
	}
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
	for k := range set {
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
	responders := p.getProvidersResponders(push)
	var wg sync.WaitGroup
	rsp := &Response{ProjectInvalidations: make(map[string]*DeviceIdList)}

	var mlock sync.Mutex
	wg.Add(len(responders))
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

	p.deliverPush(push, responders)

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
