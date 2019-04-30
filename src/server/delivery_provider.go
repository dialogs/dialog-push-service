package main

import (
	"fmt"

	raven "github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type PushTask struct {
	deviceIds     []string
	body          *PushBody
	responder     Responder
	correlationId string
}

type PushResult struct {
	ProjectId string
	Failures  *DeviceIdList
}

type DeliveryProvider interface {
	getWorkerName() string
	getTasksChan() chan PushTask
	spawnWorker(string, *providerMetrics)
	getWorkersPool() workersPool
	shouldInvalidate(string) bool
}

func spawnWorkers(d DeliveryProvider, pm *providerMetrics) {
	for i := 0; i < int(d.getWorkersPool().Workers); i++ {
		workerName := fmt.Sprintf("%s.%d", d.getWorkerName(), i)
		go raven.CapturePanic(func() {
			d.spawnWorker(workerName, pm)
		}, map[string]string{"worker": workerName})
	}
}

func (p PushingServerImpl) deliverPush(push *Push, responder Responder) int {
	tasks := 0
	for projectId, deviceList := range push.Destinations {
		deviceIds := deviceList.GetDeviceIds()
		provider, exists := p.providers[projectId]
		if !exists {
			log.WithField("correlationId", push.CorrelationId).Errorf("No provider found for projectId: %s", projectId)
			continue
		}
		if len(deviceIds) == 0 {
			log.WithField("correlationId", push.CorrelationId).Infof("Empty deviceIds: %s", push.CorrelationId)
			continue
		}
		if len(deviceIds) >= 1000 {
			log.WithField("correlationId", push.CorrelationId).Warnf("DeviceIds should be at most 999 items long %s", push.CorrelationId)
			continue
		}
		provider.getTasksChan() <- PushTask{deviceIds: deviceIds, body: push.GetBody(), responder: responder, correlationId: push.CorrelationId}
		tasks++
	}
	return tasks
}
