package server

import (
	"github.com/edganiukov/fcm"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"strings"
	"go.uber.org/zap"
)

var fcmIOHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{Namespace: "google", Name: "fcm_io", Help: "Time spent in interactions with FCM"})

type GoogleDeliveryProvider struct {
	tasks  chan PushTask
	config googleConfig
	logger *zap.Logger
}

var ErrInvalidRegistration = fcm.ErrInvalidRegistration.Error()
var ErrNotRegistered = fcm.ErrNotRegistered.Error()

func (d GoogleDeliveryProvider) getWorkerName() string {
	return d.config.ProjectID
}

func (d GoogleDeliveryProvider) shouldInvalidate(err string) bool {
	return err == ErrInvalidRegistration || err == ErrNotRegistered
}

func (d GoogleDeliveryProvider) populateFcmMessage(msg *fcm.Message, task PushTask) {
	msg.RegistrationIDs = task.deviceIds
	if voip := task.body.GetVoipPush(); voip != nil {
		d.logger.Warn("VOIP pushes are not supported, sending silent push instead")
		msg.Data["callId"] = voip.GetCallId()
		msg.Data["attemptIndex"] = voip.GetAttemptIndex()
	}
	if alerting := task.body.GetAlertingPush(); alerting != nil {
		d.logger.Warn("Alerting pushes are not supported for FCM, sending silent push instead")
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

//var stringsArrayMarshaller = zapcore.PrimitiveArrayEncoder()

func (d GoogleDeliveryProvider) spawnWorker(workerName string) {
	var err error
	var task PushTask
	msg := &fcm.Message{Data: make(map[string]interface{}), Priority: "high", DryRun: d.config.IsSandbox}
	var resp *fcm.Response
	client, err := d.getClient()
	workerLogger := d.logger.With(zap.String("worker", workerName), zap.String("key", d.config.Key))
	if err != nil {
		workerLogger.Error("Error in spawning FCM worker", zap.Error(err))
		return
	}
	subsystemName := strings.Replace(workerName, ".", "_", -1)
	successCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "processed_tasks", Help: "Tasks processed by worker"})
	failsCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "failed_tasks", Help: "Failed tasks"})
	pushesSent := prometheus.NewCounter(prometheus.CounterOpts{Namespace:"google", Subsystem: subsystemName, Name: "pushes_sent", Help: "Pushes sent (w/o result checK)"})
	prometheus.MustRegister(successCount, failsCount, pushesSent)
	workerLogger.Info("Started FCM worker")
	for task = range d.getTasksChan() {
		resetFcmMessage(msg)
		d.populateFcmMessage(msg, task)
		msg.RegistrationIDs = task.deviceIds
		beforeIO := time.Now()
		resp, err = client.SendWithRetry(msg, int(d.config.Retries))
		afterIO := time.Now()
		deviceIdKey := zap.Strings("deviceId", task.deviceIds)
		if err != nil {
			workerLogger.Error("FCM response error", zap.Error(err), zap.Any("message", msg), deviceIdKey)
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
						workerLogger.Error("FCM response error", zap.String("deviceId", task.deviceIds[k]), zap.Error(err))
					}
				}
			}
			if len(failures) > 0 {
				task.resp <- failures
			}
		} else {
			workerLogger.Info("Sucessfully sent", deviceIdKey, zap.Any("body", task.body))
		}
	}
}

func (d GoogleDeliveryProvider) getTasksChan() chan PushTask {
	return d.tasks
}

func (d GoogleDeliveryProvider) getWorkersPool() workersPool {
	return d.config.workersPool
}

func (config googleConfig) newProvider(logger *zap.Logger) DeliveryProvider {
	tasks := make(chan PushTask)
	provider := GoogleDeliveryProvider{tasks: tasks, config: config, logger: logger}
	return provider
}
