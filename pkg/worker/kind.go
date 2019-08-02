package worker

import (
	"fmt"

	"github.com/dialogs/dialog-go-lib/enum"
)

const (
	KindUnknown   Kind = 0
	KindApns      Kind = 1
	KindFcm       Kind = 2
	KindFcmLegacy Kind = 3
)

type Kind int

var _KindEnum = enum.New("worker kind").
	Add(KindUnknown, "unknown").
	Add(KindApns, "apns").    // apns: https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/APNSOverview.html#//apple_ref/doc/uid/TP40008194-CH8-SW1
	Add(KindFcm, "fcm-v1").   // new fcm: https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#http-request
	Add(KindFcmLegacy, "fcm") // legacy fcm: https://firebase.google.com/docs/cloud-messaging/http-server-ref

func KindStringKeys() []string {
	return _KindEnum.StringKeys()
}

func KindByString(src string) Kind {
	mode, ok := _KindEnum.GetByString(src)
	if !ok {
		return KindUnknown
	}
	return mode.(Kind)
}

func (k Kind) String() string {
	val, ok := _KindEnum.GetByIndex(k)
	if !ok {
		return fmt.Sprintf("invalid worker kind: %d", k)
	}

	return val
}
