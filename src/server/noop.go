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
	tasks := make(chan PushTask)
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
	workerLogger.Infof("Started NOOP provider")
	var delay int
	for task := range ndp.getTasksChan() {
		delay = rand.Intn(15)
		workerLogger.Infof("Got task: %v. Waiting %d", task, delay)
		<-time.After(time.Duration(delay) * time.Microsecond)
	}
}

func (ndp NoopDeliveryProvider) getWorkersPool() workersPool {
	return ndp.config.workersPool
}

func (ndp NoopDeliveryProvider) shouldInvalidate(string) bool {
	return false
}
