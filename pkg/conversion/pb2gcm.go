package conversion

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/provider/gcm"
)

func RequestPbToGcm(in *api.PushBody, allowAlerts bool) (*gcm.Request, error) {

	var (
		out gcm.Request
		err error
	)

	data := make(map[string]interface{})
	out.Priority = "high"

	if voip := in.GetVoipPush(); voip != nil {
		setVoIPPayloadGcm(data, voip)

	} else if encrypted := in.GetEncryptedPush(); encrypted != nil {
		err = setEncryptedPushGcm(&out, data, encrypted)

	} else if alerting := in.GetAlertingPush(); alerting != nil {
		err = serAlertingPushGcm(&out, data, alerting, allowAlerts)

	} else if silent := in.GetSilentPush(); silent != nil {
		// ignoring

	} else {
		err = ErrorByIncomingMessage(in)

	}

	if err != nil {
		return nil, err
	}

	if collapseKey := in.GetCollapseKey(); len(collapseKey) > 0 {
		out.CollapseKey = collapseKey
	}

	if ttl := in.GetTimeToLive(); ttl > 0 {
		out.TimeToLive = int(ttl)
	}

	if seq := in.GetSeq(); seq > 0 {
		data["seq"] = seq
	}

	jData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	out.Data = jData

	return &out, nil
}

func setVoIPPayloadGcm(data map[string]interface{}, src *api.VoipPush) {

	// VoIP pushes are not supported, sending silent push instead

	data["callId"] = src.GetCallId()
	data["attemptIndex"] = src.GetAttemptIndex()
	data["displayName"] = src.GetDisplayName()
	data["eventBusId"] = src.GetEventBusId()
	data["updateType"] = src.GetUpdateType()
	data["disposalReason"] = src.GetDisposalReason()

	if peer := src.GetPeer(); peer != nil {
		data["peer"] = map[string]string{
			"id":    strconv.Itoa(int(peer.Id)),
			"type":  strconv.Itoa(PeerTypeProtobufToMPS(peer.Type)),
			"strId": peer.StrId}
	}

	if outPeer := src.GetOutPeer(); outPeer != nil {
		data["outPeer"] = map[string]string{
			"id":         strconv.Itoa(int(outPeer.Id)),
			"type":       strconv.Itoa(PeerTypeProtobufToMPS(outPeer.Type)),
			"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
			"strId":      outPeer.StrId}
	}

	data["video"] = src.GetVideo()
}

func setEncryptedPushGcm(req *gcm.Request, data map[string]interface{}, src *api.EncryptedPush) error {

	if public := src.GetPublicAlertingPush(); public != nil {
		if err := setNotificationPropsGcm(req, public); err != nil {
			return err
		}
	}

	encryptedData := src.GetEncryptedData()
	if len(encryptedData) == 0 {
		return ErrEmptyEncryptedPayload
	}

	userInfo := map[string]string{
		"nonce":     strconv.Itoa(int(src.Nonce)),
		"encrypted": base64.StdEncoding.EncodeToString(encryptedData),
	}

	data["userInfo"] = userInfo

	return nil
}

func serAlertingPushGcm(req *gcm.Request, data map[string]interface{}, src *api.AlertingPush, allowAlerts bool) error {

	if !allowAlerts {
		return ErrNotSupportedAlertPush
	}

	if err := setNotificationPropsGcm(req, src); err != nil {
		return err
	}

	if mid := src.Mid; mid != nil {
		data["mid"] = mid.Value
	}

	if category := src.Category; category != nil {
		data["category"] = category.Value
	}

	return nil
}

func setNotificationPropsGcm(req *gcm.Request, src *api.AlertingPush) error {

	n := &gcm.Notification{
		Title: src.GetSimpleAlertTitle(),
		Body:  src.GetSimpleAlertBody(),
	}

	// src.GetBadge() is not supported

	data, err := n.MarshalJSON()
	if err != nil {
		return err
	}

	req.Notification = data
	return nil
}
