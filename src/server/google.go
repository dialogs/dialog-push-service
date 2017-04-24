package server

import (
	"github.com/edganiukov/fcm"
	"google.golang.org/grpc/grpclog"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"strings"
)

var fcmIOHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{Namespace: "google", Name: "fcm_io", Help: "Time spent in interactions with FCM"})

type GoogleDeliveryProvider struct {
	tasks  chan PushTask
	config googleConfig
}

var ErrInvalidRegistration = fcm.ErrInvalidRegistration.Error()
var ErrNotRegistered = fcm.ErrNotRegistered.Error()

func (d GoogleDeliveryProvider) getWorkerName() string {
	return d.config.ProjectID
}

func (d GoogleDeliveryProvider) shouldInvalidate(err string) bool {
	return err == ErrInvalidRegistration || err == ErrNotRegistered
}

func populateFcmMessage(msg *fcm.Message, task PushTask) {
	msg.RegistrationIDs = task.deviceIds
	if voip := task.body.GetVoipPush(); voip != nil {
		grpclog.Printf("VOIP pushes are not supported, sending silent push instead")
		msg.Data["callId"] = voip.GetCallId()
		msg.Data["attemptIndex"] = voip.GetAttemptIndex()
	}
	if alerting := task.body.GetAlertingPush(); alerting != nil {
		grpclog.Print("Alerting pushes are not supported for FCM, sending silent push instead")
	}
	if collapseKey := task.body.GetCollapseKey(); len(collapseKey) > 0 {
		msg.CollapseKey = collapseKey
	}
	if ttl := task.body.GetTimeToLive(); ttl > 0 {
		msg.TimeToLive = int(ttl)
	}
	if seq := task.body.GetSeq(); seq > 0 {
		msg.Data["seq"] = seq
	}
}

func resetFcmMessage(msg *fcm.Message) {
	for k := range msg.Data {
		delete(msg.Data, k)
	}
	msg.RegistrationIDs = msg.RegistrationIDs[:0]
	msg.CollapseKey = ""
	msg.TimeToLive = 0
}

func (d GoogleDeliveryProvider) getClient() (*fcm.Client, error) {
	var endpoint string
	if len(d.config.host) > 0 {
		endpoint = d.config.host
	} else {
		endpoint = fcm.DefaultEndpoint
	}
	client, err := fcm.NewClient(d.config.Key, fcm.WithEndpoint(endpoint))
	return client, err
}

func (d GoogleDeliveryProvider) spawnWorker(workerName string) {
	var err error
	var task PushTask
	msg := &fcm.Message{Data: make(map[string]interface{}), Priority: "high", DryRun: d.config.IsSandbox}
	var resp *fcm.Response
	client, err := d.getClient()
	if err != nil {
		grpclog.Printf("Error in spawning FCM worker %s: %s", workerName, err.Error())
		return
	}
	subsystemName := strings.Replace(workerName, ".", "_", -1)
	successCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "processed_tasks", Help: "Tasks processed by worker"})
	failsCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "failed_tasks", Help: "Failed tasks"})
	pushesSent := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "pushes_sent", Help: "Pushes sent (w/o result checK)"})
	prometheus.MustRegister(successCount, failsCount, pushesSent)
	grpclog.Printf("Started FCM worker %s", workerName)
	for task = range d.getTasksChan() {
		populateFcmMessage(msg, task)
		msg.RegistrationIDs = task.deviceIds
		beforeIO := time.Now()
		resp, err = client.SendWithRetry(msg, int(d.config.Retries))
		afterIO := time.Now()
		resetFcmMessage(msg)
		if err != nil {
			grpclog.Printf("[%s] FCM response error: `%s`", workerName, err.Error())
			failsCount.Inc()
			continue
		} else {
			successCount.Inc()
			fcmIOHistogram.Observe(afterIO.Sub(beforeIO).Seconds())
			pushesSent.Add(float64(len(task.deviceIds)))
		}
		if resp.Failure > 0 {
			failures := make([]string, 0, len(task.deviceIds))
			for k, r := range resp.Results {
				if r.Error != nil {
					if d.shouldInvalidate(r.Error.Error()) {
						failures = append(failures, task.deviceIds[k])
					} else {
						grpclog.Printf("[%s] FCM response error: `%s`", workerName, r.Error.Error())
					}
				}
			}
			if len(failures) > 0 {
				task.resp <- failures
			}
		} else {
			grpclog.Printf("[%s] Sucessfully sent to %s", workerName, task.deviceIds)
		}
	}
}

func (d GoogleDeliveryProvider) getTasksChan() chan PushTask {
	return d.tasks
}

func (d GoogleDeliveryProvider) getWorkersPool() workersPool {
	return d.config.workersPool
}

func (config googleConfig) newProvider() DeliveryProvider {
	tasks := make(chan PushTask)
	provider := GoogleDeliveryProvider{tasks: tasks, config: config}
	return provider
}
