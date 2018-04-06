package main

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	raven "github.com/getsentry/raven-go"
	apns "github.com/sideshow/apns2"
	pl "github.com/sideshow/apns2/payload"
	log "github.com/sirupsen/logrus"
)

type APNSDeliveryProvider struct {
	tasks  chan PushTask
	cert   tls.Certificate
	config apnsConfig
}

func (d APNSDeliveryProvider) getWorkerName() string {
	return d.config.ProjectID
}

func (d APNSDeliveryProvider) getClient() *apns.Client {
	client := apns.NewClient(d.cert)
	switch {
	case d.config.IsSandbox:
		client.Development()
	default:
		client.Production()
	}
	return client
}

func (d APNSDeliveryProvider) getTasksChan() chan PushTask {
	return d.tasks
}

func parsePrivateKey(bytes []byte) (crypto.PrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func decryptPemBlock(block *pem.Block) (crypto.PrivateKey, error) {
	if x509.IsEncryptedPEMBlock(block) {
		bytes, err := x509.DecryptPEMBlock(block, []byte(""))
		if err != nil {
			return nil, err
		}
		return parsePrivateKey(bytes)
	}
	return parsePrivateKey(block.Bytes)
}

func loadCertificate(filename string) (cert tls.Certificate, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	var block *pem.Block
	for {
		block, bytes = pem.Decode(bytes)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		}
		if block.Type == "RSA PRIVATE KEY" {
			cert.PrivateKey, err = decryptPemBlock(block)
			if err != nil {
				return
			}
		}
		if block.Type == "PRIVATE KEY" {
			cert.PrivateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return
			}
		}
	}
	if cert.PrivateKey == nil {
		err = fmt.Errorf("Private key was not extracted from %s", filename)
		return
	}
	return
}

func apnsFromAlerting(payload *pl.Payload, alerting *AlertingPush, sound string) *pl.Payload {
	if locAlert := alerting.GetLocAlertTitle(); locAlert != nil {
		payload.AlertTitleLocKey(locAlert.GetLocKey())
		payload.AlertTitleLocArgs(locAlert.GetLocArgs())
	} else if simpleTitle := alerting.GetSimpleAlertTitle(); len(simpleTitle) > 0 {
		payload.AlertTitle(simpleTitle)
	}
	if locBody := alerting.GetLocAlertBody(); locBody != nil {
		payload.AlertLocKey(locBody.GetLocKey())
		payload.AlertLocArgs(locBody.GetLocArgs())
	} else if simpleBody := alerting.GetSimpleAlertBody(); len(simpleBody) > 0 {
		payload.AlertBody(simpleBody)
	}
	if len(sound) > 0 {
		payload.Sound(sound)
	}
	if badge := alerting.GetBadge(); badge > 0 {
		payload.Badge(int(badge))
	}
	return payload
}

func (d APNSDeliveryProvider) getPayload(task PushTask, logger *log.Entry) *pl.Payload {
	// TODO: sync.Pool this?
	payload := pl.NewPayload()
	if voip := task.body.GetVoipPush(); voip != nil {
		if !d.config.IsVoip {
			logger.Warn("Attempted voip-push using non-voip certificate")
			return nil
		}
		payload.Custom("callId", voip.GetCallId())
		payload.Custom("attemptIndex", voip.GetAttemptIndex())
		payload.Custom("displayName", voip.GetDisplayName())
		payload.Custom("eventBusId", voip.GetEventBusId())
		payload.Custom("updateType", voip.GetUpdateType())
		payload.Custom("disposalReason", voip.GetDisposalReason())
		if peer := voip.GetPeer(); peer != nil {
			peerMap := map[string]string{
				"id":    strconv.Itoa(int(peer.Id)),
				"type":  strconv.Itoa(peerTypeProtobufToMPS(peer.Type)),
				"strId": peer.StrId}
			payload.Custom("peer", peerMap)
		}
		if outPeer := voip.GetOutPeer(); outPeer != nil {
			peerMap := map[string]string{
				"id":         strconv.Itoa(int(outPeer.Id)),
				"type":       strconv.Itoa(peerTypeProtobufToMPS(outPeer.Type)),
				"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
				"strId":      outPeer.StrId}
			payload.Custom("outPeer", peerMap)
		}
		payload.Custom("video", voip.GetVideo())
	}
	if alerting := task.body.GetAlertingPush(); alerting != nil {
		if d.config.IsVoip {
			logger.Warn("Attempted non-voip using voip certificate")
			return nil
		}
		if !d.config.AllowAlerts {
			logger.Warn("Alerting pushes are disabled, sending silent instead")
			if badge := alerting.GetBadge(); badge > 0 {
				payload.Badge(int(badge))
			}
			payload.ContentAvailable()
			payload.Sound("")
		} else {
			payload = apnsFromAlerting(payload, alerting, d.config.Sound)
			payload.MutableContent()
		}
	}
	if encryped := task.body.GetEncryptedPush(); encryped != nil {
		if public := encryped.GetPublicAlertingPush(); public != nil {
			payload = apnsFromAlerting(payload, public, d.config.Sound)
		}
		userInfo := make(map[string]string)
		userInfo["nonce"] = strconv.Itoa(int(encryped.Nonce))
		if data := encryped.GetEncryptedData(); data != nil && len(data) > 0 {
			userInfo["encrypted_data"] = base64.StdEncoding.EncodeToString(data)
		} else {
			logger.Warn("Encrypted push without encrypted data, ignoring")
			return nil
		}
		payload.MutableContent()
		payload.Custom("user_info", userInfo)
		// if encoded, err := json.Marshal(userInfo); err == nil {
		// 	payload.MutableContent()
		// 	payload.Custom("user_info", base64.StdEncoding.EncodeToString(encoded))
		// } else {
		// 	d.logger.Warn("Failed to marshal", zap.Error(err), zap.Any("user_info", userInfo))
		// }
	}
	if silent := task.body.GetSilentPush(); silent != nil {
		logger.Warn("Ignoring silent push")
		return nil
		/*
			if d.config.IsVoip {
				d.logger.Warn("Attempted non-voip using voip certificate")
				return nil
			}
			payload.ContentAvailable()
			payload.Sound("")
		*/
	}
	if seq := task.body.GetSeq(); seq > 0 {
		payload.Custom("seq", seq)
	}
	return payload
}

func (d APNSDeliveryProvider) getPushStatus() string {
	if d.config.AllowAlerts {
		return "alerts allowed"
	} else {
		return "silent-only"
	}
}

func (d APNSDeliveryProvider) spawnWorker(workerName string, pm *providerMetrics) {
	var err error
	var resp *apns.Response
	// TODO: there is no need in constant reallocations of pl.Payload, the allocated instance should be reused
	var payload *pl.Payload
	var task PushTask
	client := d.getClient()
	workerLogger := log.NewEntry(log.StandardLogger()).WithField("worker", workerName)
	workerLogger.Infof("Started APNS worker (%s, sound=%s)", d.getPushStatus(), d.config.Sound)
	for task = range d.getTasksChan() {
		taskLogger := workerLogger.WithField("correlationId", task.correlationId)
		// TODO: avoid allocation here, reuse payload across requests
		n := &apns.Notification{}
		payload = d.getPayload(task, taskLogger)
		if payload == nil {
			continue
		}
		taskLogger.Infof("Push transformation: `%+v` to `%+v`", task.body, payload)
		/*
			if task.body.TimeToLive > 0 {
				n.Expiration = time.Now().Add(task.body.TimeToLive * time.Second)
			}
		*/
		n.Expiration = time.Now().Add(20 * time.Minute)
		n.CollapseID = task.body.GetCollapseKey()
		n.Topic = d.config.Topic
		n.Payload = payload
		failures := make([]string, 0, len(task.deviceIds))
		for _, deviceID := range task.deviceIds {
			deviceLogger := taskLogger.WithField("deviceId", deviceID)
			deviceLogger.Info("Sending push")
			n.DeviceToken = deviceID
			beforeIO := time.Now()
			resp, err = client.Push(n)
			afterIO := time.Now()
			if err != nil {
				deviceLogger.Errorf("APNS send error %s", err.Error())
				raven.CaptureError(err, map[string]string{"deviceId": deviceID, "projectId": d.config.ProjectID})
				//metrics.fails.
				continue
			} else {
				pm.io.Observe(afterIO.Sub(beforeIO).Seconds())
				pm.success.Inc()
			}
			if !resp.Sent() {
				if d.shouldInvalidate(resp.Reason) {
					deviceLogger.Warnf("Invalidating token because of APNS response. Reason: %s", resp.Reason)
					failures = append(failures, deviceID)
				} else {
					deviceLogger.Warnf("APNS send error. Reason = %s (status = %d)", resp.Reason, resp.StatusCode)
				}
			} else {
				deviceLogger.Info("Sucessfully sent")
			}
		}
		pm.pushes.Add(float64(len(task.deviceIds)))
		if len(failures) > 0 {
			task.resp <- failures
		}
	}
}

func (d APNSDeliveryProvider) shouldInvalidate(res string) bool {
	return res == apns.ReasonBadDeviceToken ||
		res == apns.ReasonUnregistered ||
		res == apns.ReasonMissingDeviceToken ||
		res == apns.ReasonDeviceTokenNotForTopic
}

func (d APNSDeliveryProvider) getWorkersPool() workersPool {
	return d.config.workersPool
}

func (config apnsConfig) newProvider() DeliveryProvider {
	tasks := make(chan PushTask)
	cert, err := loadCertificate(config.PemFile)
	if err != nil {
		log.Fatalf("Cannot start APNS provider: %s", err.Error())
	}
	return APNSDeliveryProvider{tasks: tasks, cert: cert, config: config}
}
