package api2legacyfcm

import (
	"encoding/base64"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/edganiukov/fcm"
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

	req, ok := out.(*fcm.Message)
	if !ok {
		return converter.ErrInvalidOutgoingDataType
	}

	req.Data = make(map[string]interface{})
	req.Priority = "high"
	req.DryRun = r.sandbox

	if voip := body.GetVoipPush(); voip != nil {
		r.setVoIPPayload(req, voip)

	} else if encrypted := body.GetEncryptedPush(); encrypted != nil {
		err = r.serEncryptedPush(req, encrypted)

	} else if alerting := body.GetAlertingPush(); alerting != nil {
		err = r.serAlertingPush(req, alerting)

	} else {
		err = converter.ErrInvalidIncomingPayloadData
	}

	if err != nil {
		return err
	}

	if collapseKey := body.GetCollapseKey(); len(collapseKey) > 0 {
		req.CollapseKey = collapseKey
	}

	if ttl := body.GetTimeToLive(); ttl > 0 {
		req.TimeToLive = int(ttl)
	}

	if seq := body.GetSeq(); seq > 0 {
		req.Data["seq"] = seq
	}

	return nil
}

func (r *Request) setVoIPPayload(req *fcm.Message, src *api.VoipPush) {

	// VoIP pushes are not supported, sending silent push instead

	req.Data["callId"] = src.GetCallId()
	req.Data["attemptIndex"] = src.GetAttemptIndex()
	req.Data["displayName"] = src.GetDisplayName()
	req.Data["eventBusId"] = src.GetEventBusId()
	req.Data["updateType"] = src.GetUpdateType()
	req.Data["disposalReason"] = src.GetDisposalReason()

	if peer := src.GetPeer(); peer != nil {
		req.Data["peer"] = map[string]string{
			"id":    strconv.Itoa(int(peer.Id)),
			"type":  strconv.Itoa(converter.PeerTypeProtobufToMPS(peer.Type)),
			"strId": peer.StrId}
	}

	if outPeer := src.GetOutPeer(); outPeer != nil {
		req.Data["outPeer"] = map[string]string{
			"id":         strconv.Itoa(int(outPeer.Id)),
			"type":       strconv.Itoa(converter.PeerTypeProtobufToMPS(outPeer.Type)),
			"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
			"strId":      outPeer.StrId}
	}

	req.Data["video"] = src.GetVideo()
}

func (r *Request) serEncryptedPush(req *fcm.Message, src *api.EncryptedPush) error {

	if public := src.GetPublicAlertingPush(); public != nil {
		fromAlertingPush(req.Notification, public)
	}

	encryptedData := src.GetEncryptedData()
	if len(encryptedData) == 0 {
		return converter.ErrEmptyEncryptedPayload
	}

	userInfo := map[string]string{
		"nonce":     strconv.Itoa(int(src.Nonce)),
		"encrypted": base64.StdEncoding.EncodeToString(encryptedData),
	}

	req.Data["userInfo"] = userInfo

	return nil
}

func (r *Request) serAlertingPush(req *fcm.Message, src *api.AlertingPush) error {

	if !r.allowAlerts {
		return converter.ErrNotSupportedAlertPush
	}

	fromAlertingPush(req.Notification, src)

	if mid := src.Mid; mid != nil {
		req.Data["mid"] = mid.Value
	}

	if category := src.Category; category != nil {
		req.Data["category"] = category.Value
	}

	return nil
}

func fromAlertingPush(req *fcm.Notification, src *api.AlertingPush) {

	req.Title = src.GetSimpleAlertTitle()
	req.Body = src.GetSimpleAlertBody()

	if badge := src.GetBadge(); badge > 0 {
		req.Badge = strconv.Itoa(int(badge))
	}
}
