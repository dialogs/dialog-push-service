package main

import (
	"fmt"

	raven "github.com/getsentry/raven-go"
)

type PushTask struct {
	deviceIds     []string
	body          *PushBody
	resp          chan []string
	correlationId string
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
		go raven.CapturePanic(func() {
			d.spawnWorker(fmt.Sprintf("%s.%d", d.getWorkerName(), i), pm)
		}, map[string]string{"worker": d.getWorkerName()})
	}
}
