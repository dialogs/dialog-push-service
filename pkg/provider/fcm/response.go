package fcm

import (
	"encoding/json"
	"strconv"
	"strings"
)

const (
	ErrorCodeUnregistered    ErrorCode = "UNREGISTERED"
	ErrorCodeUnavailable     ErrorCode = "UNAVAILABLE"
	ErrorCodeInternal        ErrorCode = "INTERNAL"
	ErrorCodeUnspecified     ErrorCode = "UNSPECIFIED_ERROR"
	ErrorCodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
)

// ErrorCode values
// https://firebase.google.com/docs/reference/fcm/rest/v1/ErrorCode
type ErrorCode string

// Response format:
// error example:
// {
//   "error": {
//     "code": 400,
//     "message": "Invalid JSON payload received. Unknown name \"\": Root element must be a message.",
//     "status": "INVALID_ARGUMENT",
//     "details": [
//       {
//         "@type": "type.googleapis.com/google.rpc.BadRequest",
//         "fieldViolations": [
//           {
//             "description": "Invalid JSON payload received. Unknown name \"\": Root element must be a message."
//           }
//         ]
//       }
//     ]
//   }
// }
//
// success example:
// {
//   "name": "projects/<project-id>/messages/0:1564476468894369%30820c6b30820c6b"
// }
type Response struct {
	Name       string     `json:"name,omitempty"`
	StatusCode int        `json:"-"`
	Error      *SendError `json:"error,omitempty"`
}

// Ok returns true if notification success send
func (r *Response) Ok() bool {
	return r != nil && r.Error == nil && len(r.Name) > 0
}

// SendError format:
// error response example:
// {
//   "code": 400,
//   "message": "Invalid JSON payload received. Unknown name \"\": Root element must be a message.",
//   "status": "INVALID_ARGUMENT",
//   "details": [
//     {
//     "@type": "type.googleapis.com/google.rpc.BadRequest",
//     "fieldViolations": [
//       {
//       "description": "Invalid JSON payload received. Unknown name \"\": Root element must be a message."
//       }
//     ]
//     }
//   ]
// }
type SendError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	// error status:
	// https://firebase.google.com/docs/reference/fcm/rest/v1/ErrorCode
	Status ErrorCode `json:"status"`

	Details json.RawMessage `json:"details"`
}

// Error is 'error' interface implementation
func (e SendError) Error() string {

	b := strings.Builder{}
	b.WriteString(strconv.Itoa(e.Code))
	b.WriteByte(' ')
	b.WriteString(e.Message)
	b.WriteByte('(')
	b.WriteString(string(e.Status))
	b.WriteByte(')')

	return b.String()
}
