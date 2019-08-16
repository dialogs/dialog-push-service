package ans

import (
	"time"
)

// Response format
// Table 8-3 APNs response headers, Table 8-5 APNs JSON data keys -
// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
type Response struct {
	ID         string       `json:"id"`
	StatusCode int          `json:"status_code"`
	Body       ResponseBody `json:"body"`
}

// Table 8-5APNs JSON data keys
// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
type ResponseBody struct {
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`
}

func NewResponse(id string, statusCode int) *Response {
	return &Response{
		ID:         id,
		StatusCode: statusCode,
	}
}

func (r *ResponseBody) GetTimestamp() time.Time {

	if r == nil || r.Timestamp == 0 {
		return time.Time{}
	}

	ms := time.Millisecond * time.Duration(r.Timestamp)

	return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Add(ms)
}
