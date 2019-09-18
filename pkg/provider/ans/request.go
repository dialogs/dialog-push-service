package ans

import (
	"encoding/json"
	"net/url"
	"time"
)

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

type Request struct {
	Token   string        `json:"token,omitempty"`
	Headers RequestHeader `json:"headers,omitempty"`

	// Payload format:
	// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/PayloadKeyReference.html#//apple_ref/doc/uid/TP40008194-CH17-SW1
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (r *Request) SetToken(token string) {
	if r != nil {
		r.Token = url.QueryEscape(token)
	}
}

func (r *Request) ShouldIgnore() bool {
	return r == nil
}
