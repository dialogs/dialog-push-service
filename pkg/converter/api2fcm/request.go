package api2fcm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/provider/fcm"
)

type Request struct {
	sandbox     bool
	allowAlerts bool
}

func NewRequestConverter(cfg *Config) *Request {
	return &Request{
		sandbox:     cfg.Sandbox,
		allowAlerts: cfg.AllowAlerts,
	}
}

func (r *Request) Convert(in interface{}, out interface{}) error {

	body, err := converter.GetAPIPushBody(in)
	if err != nil {
		return err
	}

	req, ok := out.(*fcm.Request)
	if !ok {
		return converter.ErrInvalidOutgoingDataType
	}

	req.Message.Data = map[string]string{}
	req.Message.Android = &fcm.AndroidConfig{}

	if voip := body.GetVoipPush(); voip != nil {
		err = r.setVoIPPayload(req, voip)

	} else if encrypted := body.GetEncryptedPush(); encrypted != nil {
		err = r.serEncryptedPush(req, encrypted)

	} else if alerting := body.GetAlertingPush(); alerting != nil {
		err = r.serAlertingPush(req, alerting)

	} else {
		err = converter.ErrorByIncomingMessage(body)

	}

	if err != nil {
		return err
	}

	req.ValidateOnly = r.sandbox
	req.Message.Android.Priority = fcm.AndroidMessagePriorityHigh

	if collapseKey := body.GetCollapseKey(); len(collapseKey) > 0 {
		req.Message.Android.CollapseKey = collapseKey
	}

	if ttl := body.GetTimeToLive(); ttl > 0 {
		req.Message.Android.TTL = strconv.FormatInt(int64(ttl), 10) + "s"
	}

	if seq := body.GetSeq(); seq > 0 {
		req.Message.Data["seq"] = strconv.FormatInt(int64(seq), 10)
	}

	return nil
}

func (r *Request) setVoIPPayload(req *fcm.Request, src *api.VoipPush) error {

	req.Message.Data = map[string]string{
		"callId":         strconv.FormatInt(src.GetCallId(), 10),
		"attemptIndex":   strconv.FormatInt(int64(src.GetAttemptIndex()), 10),
		"displayName":    src.GetDisplayName(),
		"eventBusId":     src.GetEventBusId(),
		"updateType":     src.GetUpdateType(),
		"disposalReason": src.GetDisposalReason(),
		"video":          strconv.FormatBool(src.GetVideo()),
	}

	if peer := src.GetPeer(); peer != nil {
		peerInfo := map[string]string{
			"id":    strconv.Itoa(int(peer.Id)),
			"type":  strconv.Itoa(converter.PeerTypeProtobufToMPS(peer.Type)),
			"strId": peer.StrId,
		}

		if err := addToMap(req.Message.Data, "peer", peerInfo); err != nil {
			return err
		}
	}

	if outPeer := src.GetOutPeer(); outPeer != nil {
		peerInfo := map[string]string{
			"id":         strconv.Itoa(int(outPeer.Id)),
			"type":       strconv.Itoa(converter.PeerTypeProtobufToMPS(outPeer.Type)),
			"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
			"strId":      outPeer.StrId,
		}

		if err := addToMap(req.Message.Data, "outPeer", peerInfo); err != nil {
			return err
		}
	}

	return nil
}

func (r *Request) serEncryptedPush(req *fcm.Request, src *api.EncryptedPush) error {

	if alerting := src.GetPublicAlertingPush(); alerting != nil {
		fromAlertingPush(req, alerting)
	}

	encryptedData := src.GetEncryptedData()
	if len(encryptedData) == 0 {
		return converter.ErrEmptyEncryptedPayload
	}

	userInfo := map[string]string{
		"nonce":     strconv.FormatInt(src.Nonce, 10),
		"encrypted": base64.StdEncoding.EncodeToString(encryptedData),
	}

	if err := addToMap(req.Message.Data, "userInfo", userInfo); err != nil {
		return err
	}

	return nil
}

func (r *Request) serAlertingPush(req *fcm.Request, src *api.AlertingPush) error {

	if !r.allowAlerts {
		return converter.ErrNotSupportedAlertPush
	}

	fromAlertingPush(req, src)

	return nil
}

func fromAlertingPush(req *fcm.Request, src *api.AlertingPush) {

	if req.Message.Notification == nil {
		req.Message.Notification = &fcm.Notification{}
	}

	req.Message.Notification.Title = src.GetSimpleAlertTitle()
	req.Message.Notification.Body = src.GetSimpleAlertBody()
	// src.GetBadge() is not supported

	if mid := src.GetMid(); mid != nil {
		req.Message.Data["mid"] = mid.Value
	}

	if category := src.GetCategory(); category != nil {
		req.Message.Data["category"] = category.Value
	}
}

func addToMap(dest map[string]string, key string, val map[string]string) error {

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(val); err != nil {
		return err
	}

	dest[key] = buf.String()

	return nil
}
