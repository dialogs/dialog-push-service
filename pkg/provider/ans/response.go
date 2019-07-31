package ans

import (
	"time"

	"github.com/sideshow/apns2"
)

// Response format
// Table 8-3 APNs response headers, Table 8-5 APNs JSON data keys -
// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
type Response struct {
	ID         string    `json:"id"`
	StatusCode int       `json:"status_code"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

func NewResponse(src *apns2.Response) *Response {
	return &Response{
		ID:         src.ApnsID,
		StatusCode: src.StatusCode,
		Reason:     src.Reason,
		Timestamp:  src.Timestamp.Time,
	}
}
