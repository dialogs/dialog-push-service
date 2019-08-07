package legacyfcm

const (
	ErrorCodeMissingRegistration = "MissingRegistration"
	ErrorCodeInvalidRegistration = "InvalidRegistration"
	ErrorCodeUnavailable         = "Unavailable"
)

type Response struct {
	MulticastID int               `json:"multicast_id"`
	Success     int               `json:"success"`
	Failure     int               `json:"failure"`
	StatusCode  int               `json:"-"`
	Results     []*ResponseResult `json:"results"`
}

type ResponseResult struct {
	MessageID      string `json:"message_id"`
	RegistrationID string `json:"registration_id"`
	// error codes:
	// https://firebase.google.com/docs/cloud-messaging/http-server-ref#table9
	Error string `json:"error"`
}
