package conversion

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/provider/fcm"
)

func RequestPbToFcm(in *api.PushBody, allowAlerts bool) (*fcm.Message, error) {

	var (
		out fcm.Message
		err error
	)

	out.Data = map[string]string{}
	out.Android = &fcm.AndroidConfig{}

	if voip := in.GetVoipPush(); voip != nil {
		err = setVoIPPayloadFcm(&out, voip)

	} else if encrypted := in.GetEncryptedPush(); encrypted != nil {
		err = setEncryptedPushFcm(&out, encrypted)

	} else if alerting := in.GetAlertingPush(); alerting != nil {
		err = setAlertingPushFcm(&out, alerting, allowAlerts)

	} else if silent := in.GetSilentPush(); silent != nil {
		// ignoring

	} else {
		err = ErrorByIncomingMessage(in)

	}

	if err != nil {
		return nil, err
	}

	out.Android.Priority = fcm.AndroidMessagePriorityHigh

	if collapseKey := in.GetCollapseKey(); len(collapseKey) > 0 {
		out.Android.CollapseKey = collapseKey
	}

	if ttl := in.GetTimeToLive(); ttl > 0 {
		out.Android.TTL = strconv.FormatInt(int64(ttl), 10) + "s"
	}

	if seq := in.GetSeq(); seq > 0 {
		out.Data["seq"] = strconv.FormatInt(int64(seq), 10)
	}

	return &out, nil
}

func setVoIPPayloadFcm(req *fcm.Message, src *api.VoipPush) error {

	req.Data = map[string]string{
		"callId":         strconv.FormatInt(src.GetCallId(), 10),
		"callIdStr":      src.GetCallIdStr(),
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
			"type":  strconv.Itoa(PeerTypeProtobufToMPS(peer.Type)),
			"strId": peer.StrId,
		}

		if err := addMapToMap(req.Data, "peer", peerInfo); err != nil {
			return err
		}
	}

	if outPeer := src.GetOutPeer(); outPeer != nil {
		peerInfo := map[string]string{
			"id":         strconv.Itoa(int(outPeer.Id)),
			"type":       strconv.Itoa(PeerTypeProtobufToMPS(outPeer.Type)),
			"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
			"strId":      outPeer.StrId,
		}

		if err := addMapToMap(req.Data, "outPeer", peerInfo); err != nil {
			return err
		}
	}

	if merge := src.GetMerge(); merge != nil {
		mergeInfo := map[string]string{
			"key":   merge.GetKey(),
			"merge": strconv.FormatBool(merge.GetMerge()),
		}
		if err := addMapToMap(req.Data, "merge", mergeInfo); err != nil {
			return err
		}
	}

	return nil
}

func setEncryptedPushFcm(req *fcm.Message, src *api.EncryptedPush) error {

	if alerting := src.GetPublicAlertingPush(); alerting != nil {
		setCategoryPropsFcm(req, alerting)
	}

	encryptedData := src.GetEncryptedData()
	if len(encryptedData) == 0 {
		return ErrEmptyEncryptedPayload
	}

	userInfo := map[string]string{
		"nonce":     strconv.FormatInt(src.Nonce, 10),
		"encrypted": base64.StdEncoding.EncodeToString(encryptedData),
	}

	if err := addMapToMap(req.Data, "userInfo", userInfo); err != nil {
		return err
	}

	return nil
}

func setAlertingPushFcm(req *fcm.Message, src *api.AlertingPush, allowAlerts bool) error {

	if !allowAlerts {
		return ErrNotSupportedAlertPush
	}

	setNotificationPropsFcm(req, src)
	setCategoryPropsFcm(req, src)

	return nil
}

func setNotificationPropsFcm(req *fcm.Message, src *api.AlertingPush) {

	if req.Notification == nil {
		req.Notification = &fcm.Notification{}
	}

	req.Notification.Title = src.GetSimpleAlertTitle()
	req.Notification.Body = src.GetSimpleAlertBody()
	// src.GetBadge() is not supported
}

func setCategoryPropsFcm(req *fcm.Message, src *api.AlertingPush) {

	if category := src.GetCategory(); category != nil {
		req.Data["category"] = category.Value
	}
}

func addMapToMap(dest map[string]string, key string, val map[string]string) error {

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(val); err != nil {
		return err
	}

	dest[key] = buf.String()

	return nil
}
