package ans

import (
	"encoding/json"
	"time"

	"github.com/sideshow/apns2"
)

type Request struct {
	Token   string        `json:"token,omitempty"`
	Headers RequestHeader `json:"headers,omitempty"`

	// Payload format:
	// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/PayloadKeyReference.html#//apple_ref/doc/uid/TP40008194-CH17-SW1
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (r *Request) native() *apns2.Notification {
	return &apns2.Notification{
		DeviceToken: r.Token,
		Payload:     r.Payload,
		ApnsID:      r.Headers.ID,
		CollapseID:  r.Headers.CollapseID,
		Topic:       r.Headers.Topic,
		Expiration:  r.Headers.Expiration,
		Priority:    r.Headers.Priority,
	}
}

// RequestHeader format:
// Table 8-2 APNs request headers -
// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
type RequestHeader struct {
	// skip headers:
	// authorization - set in apns library

	ID         string    `json:"id,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
	Priority   int       `json:"priority,omitempty"`
	Topic      string    `json:"topic,omitempty"`
	CollapseID string    `json:"collapse-id,omitempty"`
}
