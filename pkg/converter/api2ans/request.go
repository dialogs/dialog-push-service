package api2ans

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/dialogs/dialog-push-service/pkg/converter"
	"github.com/dialogs/dialog-push-service/pkg/provider/ans"
	"github.com/pkg/errors"
	"github.com/sideshow/apns2/payload"
)

var ErrIgnoringSilentPush = errors.New("ignoring silent push")

type Request struct {
	topic       string
	sound       string
	allowAlerts bool
	existVoIP   bool
}

func NewRequestConverter(cfg *Config, cert tls.Certificate) (*Request, error) {

	existVoIP, err := ans.ExistOID(&cert, ans.OidVoIP)
	if err != nil {
		return nil, errors.Wrap(err, "check VoIP mode")
	}

	if err := checkVoIPTopicByCert(cfg.Topic, &cert); err != nil {
		return nil, err
	}

	return &Request{
		topic:       cfg.Topic,
		sound:       cfg.Sound,
		allowAlerts: cfg.AllowAlerts,
		existVoIP:   existVoIP,
	}, nil
}

func (r *Request) Convert(in interface{}, out interface{}) error {

	body, err := converter.GetAPIPushBody(in)
	if err != nil {
		return err
	}

	req, ok := out.(*ans.Request)
	if !ok {
		return converter.ErrInvalidOutgoingDataType
	}

	if silent := body.GetSilentPush(); silent != nil {
		return ErrIgnoringSilentPush
	}

	payload := payload.NewPayload()
	if voip := body.GetVoipPush(); voip != nil {
		err = r.setVoIPPayload(payload, voip)

	} else if alerting := body.GetAlertingPush(); alerting != nil {
		r.setAlertingPayload(payload, alerting)

	} else if encryped := body.GetEncryptedPush(); encryped != nil {
		err = r.setEncryptedPayload(payload, encryped)

	} else {
		err = converter.ErrorByIncomingMessage(body)

	}

	if err != nil {
		return err
	}

	if seq := body.GetSeq(); seq > 0 {
		payload.Custom("seq", seq)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}

	if req.Headers.Expiration.Truncate(time.Hour).IsZero() {
		req.Headers.Expiration = time.Now().Add(20 * time.Minute) // TODO: to settings or task.body.TimeToLive?
	}

	if id := body.GetCollapseKey(); id != "" {
		req.Headers.CollapseID = id
	}

	if r.topic != "" {
		req.Headers.Topic = r.topic
	}

	req.Payload = buf.Bytes()

	return nil
}

func (r *Request) setVoIPPayload(payload *payload.Payload, src *api.VoipPush) error {

	if !r.existVoIP {
		return errors.New("attempted voip-push using non-voip certificate")
	}

	payload.Custom("callId", strconv.Itoa(int(src.GetCallId())))
	payload.Custom("attemptIndex", src.GetAttemptIndex())
	payload.Custom("displayName", src.GetDisplayName())
	payload.Custom("eventBusId", src.GetEventBusId())
	payload.Custom("updateType", src.GetUpdateType())
	payload.Custom("disposalReason", src.GetDisposalReason())

	if peer := src.GetPeer(); peer != nil {
		peerInfo := map[string]string{
			"id":    strconv.Itoa(int(peer.Id)),
			"type":  strconv.Itoa(converter.PeerTypeProtobufToMPS(peer.Type)),
			"strId": peer.StrId,
		}
		payload.Custom("peer", peerInfo)
	}

	if outPeer := src.GetOutPeer(); outPeer != nil {
		peerInfo := map[string]string{
			"id":         strconv.Itoa(int(outPeer.Id)),
			"type":       strconv.Itoa(converter.PeerTypeProtobufToMPS(outPeer.Type)),
			"accessHash": strconv.Itoa(int(outPeer.AccessHash)),
			"strId":      outPeer.StrId,
		}
		payload.Custom("outPeer", peerInfo)
	}

	payload.Custom("video", src.GetVideo())

	return nil
}

func (r *Request) setAlertingPayload(payload *payload.Payload, src *api.AlertingPush) {

	if r.allowAlerts {
		setAlertingPayload(payload, src, r.sound)
		payload.MutableContent()

		if mid := src.Mid; mid != nil {
			payload.Custom("mid", mid.Value)
		}

		if category := src.Category; category != nil {
			payload.Custom("category", category.Value)
		}

	} else {
		// alerting pushes are disabled, sending silent instead
		if badge := src.GetBadge(); badge > 0 {
			payload.Badge(int(badge))
		}

		payload.ContentAvailable()
		payload.Sound("")

	}
}

func (r *Request) setEncryptedPayload(payload *payload.Payload, src *api.EncryptedPush) error {

	if public := src.GetPublicAlertingPush(); public != nil {
		setAlertingPayload(payload, public, r.sound)
	}

	encryptedData := src.GetEncryptedData()
	if len(encryptedData) == 0 {
		return converter.ErrEmptyEncryptedPayload
	}

	userInfo := map[string]string{
		"nonce":          strconv.Itoa(int(src.Nonce)),
		"encrypted_data": base64.StdEncoding.EncodeToString(encryptedData),
	}

	payload.MutableContent()
	payload.Custom("user_info", userInfo)

	return nil
}

func setAlertingPayload(payload *payload.Payload, alerting *api.AlertingPush, sound string) {

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
}

func checkVoIPTopicByCert(topic string, cert *tls.Certificate) error {

	oidValues, err := ans.GetOIDValue(cert, ans.OidVoIPTopics)
	if err != nil {
		return errors.Wrap(err, "read VoIP topics")
	}

	topicList := make([]string, 0, 10)
	for _, value := range oidValues {
		list, err := ans.GetTopics(value)
		if err != nil {
			return err
		}

		topicList = append(topicList, list...)
	}

	// check config topic by certificate
	if topic != "" {
		var exist bool
		for i := range topicList {
			if strings.Compare(topicList[i], topic) == 0 {
				exist = true
			}
		}

		if !exist {
			return fmt.Errorf("invalid VoIP topic: '%s' (topics in certificate: %v)", topic, topicList)
		}
	}

	return nil
}
