package main

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type NoopDeliveryProvider struct {
	tasks  chan PushTask
	config noopConfig
}

func (config noopConfig) newProvider() DeliveryProvider {
	tasks := make(chan PushTask, 1)
	provider := NoopDeliveryProvider{tasks: tasks, config: config}
	return provider
}

func (ndp NoopDeliveryProvider) getWorkerName() string {
	return ndp.config.ProjectID
}

func (ndp NoopDeliveryProvider) getTasksChan() chan PushTask {
	return ndp.tasks
}

func (ndp NoopDeliveryProvider) spawnWorker(workerName string, pm *providerMetrics) {
	workerLogger := log.NewEntry(log.StandardLogger()).WithField("worker", workerName)
	workerLogger.Info("Started NOOP worker")
	var delay int
	for task := range ndp.getTasksChan() {
		if ndp.config.Delay > 0 {
			delay = rand.Intn(ndp.config.Delay)
		}
		if ndp.config.OnSend != nil {
			ndp.config.OnSend(task)
		}
		go func() { task.responder.Send(ndp.config.ProjectID, &DeviceIdList{}) }()
		<-time.After(time.Duration(delay) * time.Microsecond)
	}
}

func (ndp NoopDeliveryProvider) getWorkersPool() workersPool {
	return ndp.config.workersPool
}

func (ndp NoopDeliveryProvider) shouldInvalidate(string) bool {
	return false
}
