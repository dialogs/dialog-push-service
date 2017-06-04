package main

import "fmt"

type PushTask struct {
	deviceIds []string
	body      *PushBody
	resp      chan []string
}

type DeliveryProvider interface {
	getWorkerName() string
	getTasksChan() chan PushTask
	spawnWorker(string)
	getWorkersPool() workersPool
	shouldInvalidate(string) bool
}

func spawnWorkers(d DeliveryProvider) {
	for i := 0; i < int(d.getWorkersPool().Workers); i++ {
		go d.spawnWorker(fmt.Sprintf("%s.%d", d.getWorkerName(), i))
	}
}