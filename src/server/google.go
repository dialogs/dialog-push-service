package main

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/edganiukov/fcm"
	raven "github.com/getsentry/raven-go"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

func fcmFromAlerting(n *fcm.Notification, alerting *AlertingPush) *fcm.Notification {
	n.Title = alerting.GetSimpleAlertTitle()
	n.Body = alerting.GetSimpleAlertBody()
	if badge := alerting.GetBadge(); badge > 0 {
		n.Badge = strconv.Itoa(int(badge))
	}
	return n
}

func (d GoogleDeliveryProvider) populateFcmMessage(msg *fcm.Message, task PushTask, logger *log.Entry) bool {
	msg.RegistrationIDs = task.deviceIds
	if voip := task.body.GetVoipPush(); voip != nil {
		logger.Warn("VOIP pushes are not supported, sending silent push instead")
		msg.Data["callId"] = voip.GetCallId()
		msg.Data["attemptIndex"] = voip.GetAttemptIndex()
		msg.Data["displayName"] = voip.GetDisplayName()
		msg.Data["eventBusId"] = voip.GetEventBusId()
		msg.Data["updateType"] = voip.GetUpdateType()
		msg.Data["disposalReason"] = voip.GetDisposalReason()
		if peer := voip.GetPeer(); peer != nil {
			msg.Data["peer"] = map[string]string{
				"id":    strconv.Itoa(int(peer.Id)),
				"type":  strconv.Itoa(peerTypeProtobufToMPS(peer.Type)),
				"strId": peer.StrId}
		}
		if outPeer := voip.GetOutPeer(); outPeer != nil {
			msg.Data["outPeer"] = map[string]string{
				"id":         strconv.Itoa(int(outPeer.Id)),
				"type":       strconv.Itoa(peerTypeProtobufToMPS(outPeer.Type)),
				"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
				"strId":      outPeer.StrId}
		}
	}
	if encrypted := task.body.GetEncryptedPush(); encrypted != nil {
		if public := encrypted.GetPublicAlertingPush(); public != nil {
			msg.Notification = fcmFromAlerting(msg.Notification, public)
		}
		userInfo := make(map[string]string)
		userInfo["nonce"] = strconv.Itoa(int(encrypted.Nonce))
		if data := encrypted.GetEncryptedData(); data != nil && len(data) > 0 {
			userInfo["encrypted"] = base64.StdEncoding.EncodeToString(data)
		} else {
			logger.Warn("Encrypted push without encrypted data, ignoring")
			return false
		}
		msg.Data["userInfo"] = userInfo
	}
	if alerting := task.body.GetAlertingPush(); alerting != nil {
		if !d.config.AllowAlerts {
			logger.Warn("Alerting pushes are not supported for FCM, sending silent push instead")
		} else {
			msg.Notification = fcmFromAlerting(msg.Notification, alerting)
		}
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
	if data, err := task.body.Marshal(); err != nil {
		logger.Errorf("Failed to marshall task body. Ignoring push: %s", err.Error())
		return false
	} else {
		msg.Data["body"] = base64.StdEncoding.EncodeToString(data)
	}
	return true
}

func resetFcmMessage(msg *fcm.Message) {
	for k := range msg.Data {
		delete(msg.Data, k)
	}
	msg.Notification = &fcm.Notification{}
	msg.RegistrationIDs = msg.RegistrationIDs[:0]
	msg.CollapseKey = ""
	msg.TimeToLive = 0
}

func (d GoogleDeliveryProvider) getClient() (*fcm.Client, error) {
	client, err := fcm.NewClient(d.config.Key, fcm.WithEndpoint(fcm.DefaultEndpoint))
	return client, err
}

func (d GoogleDeliveryProvider) spawnWorker(workerName string) {
	var err error
	var task PushTask
	msg := &fcm.Message{Data: make(map[string]interface{}), Priority: "high", DryRun: d.config.IsSandbox}
	var resp *fcm.Response
	client, err := d.getClient()
	workerLogger := log.NewEntry(log.StandardLogger()).WithField("worker", workerName)
	if err != nil {
		workerLogger.Errorf("Error in spawning FCM worker: %s", err.Error())
		raven.CaptureError(err, map[string]string{"projectId": d.config.ProjectID})
		return
	}
	subsystemName := strings.Replace(workerName, ".", "_", -1)
	successCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace: "google", Subsystem: subsystemName, Name: "processed_tasks", Help: "Tasks processed by worker"})
	failsCount := prometheus.NewCounter(prometheus.CounterOpts{Namespace: "google", Subsystem: subsystemName, Name: "failed_tasks", Help: "Failed tasks"})
	pushesSent := prometheus.NewCounter(prometheus.CounterOpts{Namespace: "google", Subsystem: subsystemName, Name: "pushes_sent", Help: "Pushes sent (w/o result checK)"})
	prometheus.MustRegister(successCount, failsCount, pushesSent)
	workerLogger.Info("Started FCM worker")
	for task = range d.getTasksChan() {
		taskLogger := workerLogger.WithField("id", task.correlationId)
		resetFcmMessage(msg)
		if !d.populateFcmMessage(msg, task, taskLogger) {
			continue
		}
		msg.RegistrationIDs = task.deviceIds
		taskLogger.Infof("Sending push")
		beforeIO := time.Now()
		resp, err = client.SendWithRetry(msg, int(d.config.Retries))
		afterIO := time.Now()
		if err != nil {
			taskLogger.Errorf("FCM response error: %s", err.Error())
			raven.CaptureError(err, map[string]string{"projectId": d.config.ProjectID})
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
						taskLogger.Errorf("FCM response error for deviceId = %s: %s", task.deviceIds[k], err.Error())
					}
				}
			}
			if len(failures) > 0 {
				task.resp <- failures
			}
		} else {
			taskLogger.Info("Sucessfully sent")
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
